#!/bin/bash

function common_hooks::certificates::order_certificate::config() {
  echo '
{
  "afterHelm": 5,
  "schedule": [
    {
      "crontab": "42 4 * * *"
    }
  ]
}'
}

# $1 - имя namespace, для которого надо сгенерировать сертификат
# $2 - название секрета, куда сложить сгенерированный сертификат
# $3 - common_name генерируемого сертификата
function common_hooks::certificates::order_certificate::main() {
  namespace=$1
  secret_name=$2
  common_name=$3

  if kubectl -n ${namespace} get secret/${secret_name} > /dev/null 2> /dev/null ; then
    # Проверяем срок действия
    cert=$(kubectl -n ${namespace} get secret/${secret_name} -o jsonpath='{.data.tls\.crt}' | base64 -d)
    not_after=$(echo "$cert" | cfssl-certinfo -cert - | jq .not_after -r | sed 's/\([0-9]\{4\}-[0-9]\{2\}-[0-9]\{2\}\)T\([0-9]\{2\}:[0-9]\{2\}:[0-9]\{2\}\).*/\1 \2/')
    valid_for=$(expr $(date --date="$not_after" +%s) - $(date +%s))

    # За десять дней до окончания
    if [[ $valid_for -lt 864000 ]] ; then
      # Удаляем секрет, будет перезаказан ниже
      kubectl -n ${namespace} delete secret/${secret_name}
    else
      return 0
    fi
  fi

  # Удаляем CSR, если существовал раньше
  if kubectl get csr/${common_name} > /dev/null 2> /dev/null ; then
    kubectl delete csr/${common_name}
  fi

  # Генерируем CSR
  cfssl_result=$(jo CN=${common_name} key="$(jo algo=ecdsa size=256)" | cfssl genkey -)
  cfssl_result_csr=$(echo "$cfssl_result" | jq .csr -r | base64 | tr -d '\n')
  csr=$(cat <<EOF
apiVersion: certificates.k8s.io/v1beta1
kind: CertificateSigningRequest
metadata:
  name: ${common_name}
spec:
  request: ${cfssl_result_csr}
  usages:
  - digital signature
  - key encipherment
  - client auth
EOF
)

  # Создаем CSR и сразу его подтверждаем
  echo "$csr" | kubectl create -f -
  echo "$csr" | kubectl certificate approve -f -

  # Дожидаемся подписанного сертификата, скачеваем его и удаляем CSR
  for i in $(seq 1 120); do
    if [[ "$(kubectl get csr/${common_name} -o json | jq '.status | has("certificate")')" == "true" ]] ; then
      break
    fi

    echo "Wait for csr/${common_name} approval"
    sleep 1
  done
  if [[ $i -gt 120 ]] ; then
    >&2 echo "Timeout waiting for csr/${common_name} approval"
    return 1
  fi
  cert=$(kubectl get csr/${common_name} -o jsonpath='{.status.certificate}')
  kubectl delete csr/${common_name}

  # Создаем секрет
  key=$(echo "$cfssl_result" | jq .key -r | base64 | tr -d '\n')
  kubectl create -f - <<EOF
apiVersion: v1
metadata:
  name: ${secret_name}
  namespace: ${namespace}
type: kubernetes.io/tls
data:
  tls.crt: $cert
  tls.key: $key
kind: Secret
EOF
}

