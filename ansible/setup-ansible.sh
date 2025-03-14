#!/bin/bash


# install ansible
pip install ansible  --break-system-packages --no-cache-dir

# run ansible playbook
ansible-playbook -i inventory.yml deploy_vector.yml