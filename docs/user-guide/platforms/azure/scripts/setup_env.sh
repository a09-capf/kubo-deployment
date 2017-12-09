#!/usr/bin/env bash

source utils.sh

echo "Start to update package lists from repositories..."
retryop "apt-get update"

echo "Start to install prerequisites..."
retryop "apt-get -y install build-essential zlibc zlib1g-dev ruby ruby-dev openssl libxslt-dev libxml2-dev libssl-dev libreadline6 libreadline6-dev libyaml-dev libsqlite3-dev sqlite3 python-dev python-pip jq"

set -e

tenant_id=$1
client_id=$2
client_secret=$(echo $3 | base64 --decode)
custom_data_file="/var/lib/cloud/instance/user-data.txt"
settings=$(cat ${custom_data_file})

function get_setting() {
  key=$1
  local value=$(echo $settings | jq ".$key" -r)
  echo $value
}

function install_bosh_cli() {
  echo "Start to install bosh-cli v2..."
  bosh_cli_url=$1
  sudo curl $bosh_cli_url -o /usr/bin/bosh-cli
  sudo chmod a+x /usr/bin/bosh-cli
}

function install_credhub_cli() {
  echo "Start to install credhub-cli..."
  credhub_cli_url=$1
  curl -L $credhub_cli_url | tar zxv
  chmod a+x credhub
  sudo mv credhub /usr/bin
}

function install_kubectl() {
  echo "Start to install kubectl..."
  kubectl_url=$1
  sudo curl -L $kubectl_url -o /usr/bin/kubectl
  sudo chmod a+x /usr/bin/kubectl
}

environment=$(get_setting ENVIRONMENT)

set +e

echo "Start to install python packages..."
pkg_list="setuptools==32.3.1 azure==2.0.0rc1"
for pkg in $pkg_list; do
  retryop "pip install $pkg"
done

set -e

echo "Creating the containers (bosh and stemcell) and the table (stemcells) in the default storage account"
default_storage_account=$(get_setting DEFAULT_STORAGE_ACCOUNT_NAME)
default_storage_access_key=$(get_setting DEFAULT_STORAGE_ACCESS_KEY)
endpoint_suffix=$(get_setting SERVICE_HOST_BASE)
python prepare_storage_account.py ${default_storage_account} ${default_storage_access_key} ${endpoint_suffix} ${environment}

bosh_cli_url=$(get_setting BOSH_CLI_URL)
install_bosh_cli $bosh_cli_url

credhub_cli_url=$(get_setting CREDHUB_CLI_URL)
install_credhub_cli $credhub_cli_url

kubectl_url=$(get_setting KUBECTL_URL)
install_kubectl $kubectl_url

username=$(get_setting ADMIN_USER_NAME)
home_dir="/home/$username"
chown -R $username $home_dir

echo "The devbox is prepared successfully."
exit 0
