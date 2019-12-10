package hooks

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strings"

	. "github.com/flant/libjq-go"

	"github.com/tidwall/gjson"

	"github.com/deckhouse/deckhouse/testing/library/sandbox_runner"

	"github.com/deckhouse/deckhouse/testing/library/values_store"

	jsonpatch "gopkg.in/evanphx/json-patch.v4"

	"gopkg.in/yaml.v3"

	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/flant/shell-operator/test/hook/context"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/deckhouse/deckhouse/testing/library/object_store"
)

var globalTmpDir string

func (hec *HookExecutionConfig) KubernetesGlobalResource(kind, name string) object_store.KubeObject {
	return hec.ObjectStore.KubernetesGlobalResource(kind, name)
}

func (hec *HookExecutionConfig) KubernetesResource(kind, namespace, name string) object_store.KubeObject {
	return hec.ObjectStore.KubernetesResource(kind, namespace, name)
}

type ShellOperatorHookConfig struct {
	ConfigVersion interface{} `json:"configVersion,omitempty"`
	Schedule      interface{} `json:"kubernetes,omitempty"`
	Kubernetes    interface{} `json:"schedule,omitempty"`
}

type CustomCRD struct {
	Group      string
	Version    string
	Kind       string
	Namespaced bool
}

type HookExecutionConfig struct {
	tmpDir            string // FIXME
	HookPath          string
	values            *values_store.ValuesStore
	configValues      *values_store.ValuesStore
	hookConfig        string // <hook> --config output
	KubeExtraCRDs     []CustomCRD
	IsKubeStateInited bool
	KubeState         string // yaml string
	ObjectStore       object_store.ObjectStore
	//KubernetesClusterObjects map[string]unstructured.Unstructured
	BindingContexts          BindingContextsSlice
	BindingContextsRaw       string // array of contexts
	BindingContextController context.BindingContextController

	Session *gexec.Session
}

func (hec *HookExecutionConfig) RegisterCRD(group, version, kind string, namespaced bool) {
	newCRD := CustomCRD{Group: group, Version: version, Kind: kind, Namespaced: namespaced}
	hec.KubeExtraCRDs = append(hec.KubeExtraCRDs, newCRD)
}

func (hec *HookExecutionConfig) ValuesGet(path string) gjson.Result {
	return hec.values.Get(path)
}

func (hec *HookExecutionConfig) ConfigValuesGet(path string) gjson.Result {
	return hec.configValues.Get(path)
}

func (hec *HookExecutionConfig) ValuesSet(path string, value interface{}) {
	hec.values.SetByPath(path, value)
}

func (hec *HookExecutionConfig) ConfigValuesSet(path string, value interface{}) {
	hec.configValues.SetByPath(path, value)
}

func (hec *HookExecutionConfig) ValuesSetFromYaml(path string, value []byte) {
	hec.values.SetByPathFromYaml(path, value)
}

func (hec *HookExecutionConfig) ConfigValuesSetFromYaml(path string, value []byte) {
	hec.configValues.SetByPathFromYaml(path, value)
}

func HookExecutionConfigInit(initValues, initConfigValues string) *HookExecutionConfig {
	var err error
	hookEnvs := []string{}

	hookConfig := new(HookExecutionConfig)
	_, filepath, _, ok := runtime.Caller(1)
	if !ok {
		panic("can't execute runtime.Caller")
	}
	hookConfig.HookPath = strings.TrimSuffix(filepath, "_test.go")

	hookConfig.KubeExtraCRDs = []CustomCRD{}

	BeforeEach(func() {
		hookConfig.values, err = values_store.NewStoreFromRawYaml([]byte(initValues))
		if err != nil {
			panic(err)
		}
		hookConfig.configValues, err = values_store.NewStoreFromRawYaml([]byte(initConfigValues))
		if err != nil {
			panic(err)
		}
		hookConfig.IsKubeStateInited = false
		hookConfig.BindingContexts.Set()
	})

	hookEnvs = append(hookEnvs, "D8_IS_TESTS_ENVIRONMENT=yes")

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}
	cmd := &exec.Cmd{
		Path:   hookConfig.HookPath,
		Args:   []string{hookConfig.HookPath, "--config"},
		Env:    append(os.Environ(), hookEnvs...),
		Stdout: &stdout,
		Stderr: &stderr,
	}

	hookConfig.tmpDir, err = ioutil.TempDir(globalTmpDir, "")
	if err != nil {
		panic(err)
	}

	if err := cmd.Run(); err != nil {
		panic(err)
	}

	var config ShellOperatorHookConfig
	err = json.Unmarshal(stdout.Bytes(), &config)
	if err != nil {
		panic(err)
	}

	result, err := json.Marshal(config)
	if err != nil {
		panic(err)
	}
	hookConfig.hookConfig = string(result)

	return hookConfig
}

func (hec *HookExecutionConfig) KubeStateSet(newKubeState string) []BindingContext {
	var err error
	if hec.IsKubeStateInited == false {
		hec.BindingContextController, err = context.NewBindingContextController(hec.hookConfig, newKubeState)
		if err != nil {
			panic(err)
		}

		if len(hec.KubeExtraCRDs) > 0 {
			for _, crd := range hec.KubeExtraCRDs {
				hec.BindingContextController.RegisterCRD(crd.Group, crd.Version, crd.Kind, crd.Namespaced)
			}
		}

		hec.BindingContextsRaw, err = hec.BindingContextController.Run()
		if err != nil {
			panic(err)
		}
		hec.IsKubeStateInited = true
	} else {
		hec.BindingContextsRaw, err = hec.BindingContextController.ChangeState(newKubeState)
		if err != nil {
			panic(err)
		}
	}
	hec.KubeState = newKubeState

	var contexts []BindingContext
	err = json.Unmarshal([]byte(hec.BindingContextsRaw), &contexts)
	if err != nil {
		panic(err)
	}
	return contexts
}

func (hec *HookExecutionConfig) KubeStateToKubeObjects() error {
	var err error
	hec.ObjectStore = make(object_store.ObjectStore)
	dec := yaml.NewDecoder(strings.NewReader(hec.KubeState))
	for {
		var t interface{}
		err = dec.Decode(&t)

		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if t == nil {
			continue
		}

		var unstructuredObj unstructured.Unstructured
		unstructuredObj.SetUnstructuredContent(t.(map[string]interface{}))
		hec.ObjectStore.PutObject(unstructuredObj.Object, object_store.NewMetaIndex(unstructuredObj.GetKind(), unstructuredObj.GetNamespace(), unstructuredObj.GetName()))
	}
	return nil
}

func (hec *HookExecutionConfig) RunHook() {
	var (
		err error

		tmpDir string
		//bindingContexts []hook.BindingContextV1

		ValuesFile                *os.File
		ConfigValuesFile          *os.File
		ValuesJsonPatchFile       *os.File
		ConfigValuesJsonPatchFile *os.File
		BindingContextFile        *os.File
		KubernetesPatchSetFile    *os.File

		hookEnvs []string
	)

	err = hec.KubeStateToKubeObjects()
	Expect(err).ShouldNot(HaveOccurred())

	hookEnvs = append(hookEnvs, "D8_IS_TESTS_ENVIRONMENT=yes")

	hookCmd := &exec.Cmd{
		Path: hec.HookPath,
		Args: []string{hec.HookPath, "--config"},
		Env:  append(os.Environ(), hookEnvs...),
	}

	hec.Session, err = gexec.Start(hookCmd, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())

	<-hec.Session.Exited
	Expect(hec.Session.ExitCode()).To(Equal(0))

	out := hec.Session.Out.Contents()
	By("Parsing config " + string(out))

	var parsedConfig json.RawMessage
	Expect(json.Unmarshal(out, &parsedConfig)).To(Succeed())

	Expect(hec.values.JsonRepr).ToNot(BeEmpty())

	Expect(hec.configValues.JsonRepr).ToNot(BeEmpty())

	/* TODO: не нужно?
	if hec.BindingContextsRaw != "" {
		err := json.Unmarshal([]byte(hec.BindingContextsRaw), &bindingContexts)
		Expect(err).ShouldNot(HaveOccurred())
	}
	*/

	bindingContextBytes, err := json.Marshal(hec.BindingContexts)
	Expect(err).ShouldNot(HaveOccurred())
	hec.BindingContextsRaw = string(bindingContextBytes)

	tmpDir, err = ioutil.TempDir(globalTmpDir, "")
	Expect(err).ShouldNot(HaveOccurred())

	ValuesFile, err = ioutil.TempFile(tmpDir, "")
	Expect(err).ShouldNot(HaveOccurred())
	hookEnvs = append(hookEnvs, "VALUES_PATH="+ValuesFile.Name())

	ConfigValuesFile, err = ioutil.TempFile(tmpDir, "")
	Expect(err).ShouldNot(HaveOccurred())
	hookEnvs = append(hookEnvs, "CONFIG_VALUES_PATH="+ConfigValuesFile.Name())

	ValuesJsonPatchFile, err = ioutil.TempFile(tmpDir, "")
	Expect(err).ShouldNot(HaveOccurred())
	hookEnvs = append(hookEnvs, "VALUES_JSON_PATCH_PATH="+ValuesJsonPatchFile.Name())

	ConfigValuesJsonPatchFile, err = ioutil.TempFile(tmpDir, "")
	Expect(err).ShouldNot(HaveOccurred())
	hookEnvs = append(hookEnvs, "CONFIG_VALUES_JSON_PATCH_PATH="+ConfigValuesJsonPatchFile.Name())

	BindingContextFile, err = ioutil.TempFile(tmpDir, "")
	Expect(err).ShouldNot(HaveOccurred())
	hookEnvs = append(hookEnvs, "BINDING_CONTEXT_PATH="+BindingContextFile.Name())

	KubernetesPatchSetFile, err = ioutil.TempFile(tmpDir, "")
	Expect(err).ShouldNot(HaveOccurred())
	hookEnvs = append(hookEnvs, "D8_KUBERNETES_PATCH_SET_FILE="+KubernetesPatchSetFile.Name())

	hookCmd = &exec.Cmd{
		Path: hec.HookPath,
		Args: []string{hec.HookPath},
		Env:  hookEnvs,
	}

	hec.Session = sandbox_runner.Run(hookCmd,
		sandbox_runner.WithFile(ValuesFile.Name(), hec.values.JsonRepr),
		sandbox_runner.WithFile(ConfigValuesFile.Name(), hec.configValues.JsonRepr),
		sandbox_runner.WithFile(BindingContextFile.Name(), []byte(hec.BindingContextsRaw)),
	)

	valuesJsonPatchBytes, err := ioutil.ReadAll(ValuesJsonPatchFile)
	Expect(err).ShouldNot(HaveOccurred())
	configValuesJsonPatchBytes, err := ioutil.ReadAll(ConfigValuesJsonPatchFile)
	Expect(err).ShouldNot(HaveOccurred())
	kubernetesPatchBytes, err := ioutil.ReadAll(KubernetesPatchSetFile)
	Expect(err).ShouldNot(HaveOccurred())

	// TODO: take a closer look and refactor into a function
	if len(valuesJsonPatchBytes) != 0 {
		patch, err := jsonpatch.DecodePatch(valuesJsonPatchBytes)
		Expect(err).ShouldNot(HaveOccurred())

		patchedValuesBytes, err := patch.Apply(hec.values.JsonRepr)
		Expect(err).ShouldNot(HaveOccurred())
		hec.values = values_store.NewStoreFromRawJson(patchedValuesBytes)
	}

	if len(configValuesJsonPatchBytes) != 0 {
		patch, err := jsonpatch.DecodePatch(configValuesJsonPatchBytes)
		Expect(err).ShouldNot(HaveOccurred())

		patchedConfigValuesBytes, err := patch.Apply(hec.configValues.JsonRepr)
		Expect(err).ShouldNot(HaveOccurred())
		hec.configValues = values_store.NewStoreFromRawJson(patchedConfigValuesBytes)
	}

	if len(kubernetesPatchBytes) != 0 {
		kubePatch, err := NewKubernetesPatch(kubernetesPatchBytes)
		Expect(err).ShouldNot(HaveOccurred())

		patchedObjects, err := kubePatch.Apply(hec.ObjectStore)
		Expect(err).ToNot(HaveOccurred())

		hec.ObjectStore = patchedObjects
	}
}

var doneChan = make(chan struct{})

var _ = BeforeSuite(func() {
	go JqCallLoop(doneChan)
	By("Init temporary directories")
	var err error
	globalTmpDir, err = ioutil.TempDir("", "")
	Expect(err).ToNot(HaveOccurred())
})

var _ = AfterSuite(func() {
	By("Removing temporary directories")
	Expect(os.RemoveAll(globalTmpDir)).Should(Succeed())
	doneChan <- struct{}{}
})