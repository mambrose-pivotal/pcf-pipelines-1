package upgradeert

const PipelineYaml = `groups:
- name: ert-upgrade
  jobs:
  - upload-{{pivnet_product_name}}
  - stage-{{pivnet_product_name}}
  - stage-stemcell-{{pivnet_product_name}}
  - apply-changes-{{pivnet_product_name}}

resource_types:
- name: pivnet
  type: docker-image
  source:
    repository: pivotalcf/pivnet-resource
    tag: latest-final

resources:
- name: task-bundle-release 
  type: github-release
  source:
    user: c0-ops 
    repository: concourse-tasks-bundle 
    access_token: {{github_token}} 

- name: tool-om
  type: github-release
  source:
    user: pivotal-cf
    repository: om
    access_token: {{github_token}}

- name: pivnet-product
  type: pivnet
  check_every: {{poll_interval}} 
  source:
    api_token: {{pcf_pivnet_token}}
    product_slug: {{pivnet_product_name}}
    product_version: {{ert_major_minor_version}}
    sort_by: semver

jobs:
- name: upload-{{pivnet_product_name}}
  plan:
  - aggregate:
    - get: task-bundle-release
      params:
        globs:
        - tasks-bundle.tgz 
    - get: pivnet-product
      trigger: true
      params:
        globs:
        - "*pivotal"
    - get: tool-om
      params:
        globs:
        - om-linux

  - task: unpack-tasks
    config:
      platform: linux
      image_resource:
        type: docker-image
        source:
          repository: cloudfoundry/cflinuxfs2 
      inputs:
      - name: task-bundle-release
      outputs:
      - name: concourse-tasks-bundle 
      run:
        path: tar 
        args:
        - -C 
        - concourse-tasks-bundle
        - -xvzf
        - task-bundle-release/tasks-bundle.tgz 

  - task: upload-product
    file: concourse-tasks-bundle/upload-product/task.yml
    params:
      OPSMAN_USERNAME: {{opsman_admin_username}}
      OPSMAN_PASSWORD: {{opsman_admin_password}}
      OPSMAN_URI: {{opsman_uri}}
      PIVNET_PRODUCT_NAME: {{pivnet_product_name}}

- name: stage-{{pivnet_product_name}}
  plan:
  - aggregate:
    - get: pivnet-product
      trigger: true
      passed: [upload-{{pivnet_product_name}}]
      params:
        globs:
        - "*pivotal"
    - get: task-bundle-release
      params:
        globs:
        - tasks-bundle.tgz 
      passed: [upload-{{pivnet_product_name}}]
    - get: tool-om
      params:
        globs:
        - om-linux

  - task: unpack-tasks
    config:
      platform: linux
      image_resource:
        type: docker-image
        source:
          repository: cloudfoundry/cflinuxfs2 
      inputs:
      - name: task-bundle-release
      outputs:
      - name: concourse-tasks-bundle 
      run:
        path: tar 
        args:
        - -C 
        - concourse-tasks-bundle
        - -xvzf
        - task-bundle-release/tasks-bundle.tgz 

  - task: stage-product
    file: concourse-tasks-bundle/stage-product/task.yml
    params:
      OPSMAN_USERNAME: {{opsman_admin_username}}
      OPSMAN_PASSWORD: {{opsman_admin_password}}
      OPSMAN_URI: {{opsman_uri}}
      PRODUCT_NAME: {{opsman_product_name}}


- name: stage-stemcell-{{pivnet_product_name}}
  plan:
  - aggregate:
    - get: pivnet-product
      trigger: true
      passed: [stage-{{pivnet_product_name}}]
      params:
        globs:
        - "*pivotal"
    - get: task-bundle-release
      params:
        globs:
        - tasks-bundle.tgz 
      passed: [stage-{{pivnet_product_name}}]
    - get: tool-om
      params:
        globs:
        - om-linux

  - task: unpack-tasks
    config:
      platform: linux
      image_resource:
        type: docker-image
        source:
          repository: cloudfoundry/cflinuxfs2 
      inputs:
      - name: task-bundle-release
      outputs:
      - name: concourse-tasks-bundle 
      run:
        path: tar 
        args:
        - -C 
        - concourse-tasks-bundle
        - -xvzf
        - task-bundle-release/tasks-bundle.tgz 

  - task: download-stemcell
    file: concourse-tasks-bundle/download-bosh-io-stemcell/task.yml
    params:
      PRODUCT: {{opsman_product_name}}
      IAAS_TYPE: {{iaas_type}}

  - task: upload-stemcell
    file: concourse-tasks-bundle/upload-stemcell/task.yml
    params:
      OPSMAN_USERNAME: {{opsman_admin_username}}
      OPSMAN_PASSWORD: {{opsman_admin_password}}
      OPSMAN_URI: {{opsman_uri}}

- name: apply-changes-{{pivnet_product_name}}
  plan:
  - aggregate:
    - get: task-bundle-release
      params:
        globs:
        - tasks-bundle.tgz 
      passed: [stage-stemcell-{{pivnet_product_name}}]
    - get: tool-om
      params:
        globs:
        - om-linux

  - task: unpack-tasks
    config:
      platform: linux
      image_resource:
        type: docker-image
        source:
          repository: cloudfoundry/cflinuxfs2 
      inputs:
      - name: task-bundle-release
      outputs:
      - name: concourse-tasks-bundle 
      run:
        path: tar 
        args:
        - -C 
        - concourse-tasks-bundle
        - -xvzf
        - task-bundle-release/tasks-bundle.tgz 

  - task: apply-changes
    file: concourse-tasks-bundle/apply-changes/task.yml
    params:
      OPSMAN_USERNAME: {{opsman_admin_username}}
      OPSMAN_PASSWORD: {{opsman_admin_password}}
      OPSMAN_URI: {{opsman_uri}}
`