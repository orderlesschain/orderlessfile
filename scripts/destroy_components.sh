#!/usr/bin/env bash

#echo "Destroying the nodes and the clients ..."

source "${PROJECT_ABSOLUTE_PATH}"/env
start=$(date +%s)

if [[ $BUILD_MODE == "local" ]]; then
  pushd "${PROJECT_ABSOLUTE_PATH}"/deployment/ansible/ansible_local || exit
else
  pushd "${PROJECT_ABSOLUTE_PATH}"/deployment/ansible/ansible_remote || exit
fi


ansible-playbook "${PROJECT_ABSOLUTE_PATH}"/deployment/ansible/playbooks/destroy_node.yml
ansible-playbook "${PROJECT_ABSOLUTE_PATH}"/deployment/ansible/playbooks/destroy_client.yml

popd || exit

end=$(date +%s)

echo Built and deployed in $(expr $end - $start) seconds.
