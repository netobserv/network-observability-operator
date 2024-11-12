import os
from sys import exit as sys_exit
from datetime import datetime
from ruamel.yaml import YAML
yaml = YAML()

def load_manifest(pathn):
   if not pathn.endswith(".yaml"):
      return None
   try:
      with open(pathn, "r") as f:
         return yaml.load(f)
   except FileNotFoundError:
      print("File can not found")
      exit(6)

def dump_manifest(pathn, manifest):
   with open(pathn, "w") as f:
      yaml.dump(manifest, f)
   return

timestamp = int(os.getenv('EPOC_TIMESTAMP'))
datetime_time = datetime.fromtimestamp(timestamp)
version = os.getenv('VERSION')
replaces = os.getenv('REPLACES')
desc_file_name = os.getenv('IN_CSV_DESC')
csv = load_manifest(os.getenv('TARGET_CSV_FILE'))
created_at = datetime_time.strftime('%Y-%m-%dT%H:%M:%S')
operator_image = os.getenv('OPERATOR_IMAGE_PULLSPEC')
ebpf_image = os.getenv('EBPF_IMAGE_PULLSPEC')
flp_image = os.getenv('FLP_IMAGE_PULLSPEC')
console_image = os.getenv('CONSOLE_IMAGE_PULLSPEC')

csv['metadata']['annotations']['operators.openshift.io/valid-subscription'] = '["OpenShift Kubernetes Engine", "OpenShift Container Platform", "OpenShift Platform Plus"]'
csv['metadata']['annotations']['operatorframework.io/cluster-monitoring'] = 'true'
csv['metadata']['name'] = 'network-observability-operator.v{}'.format(version)
csv['metadata']['annotations']['createdAt'] = created_at
csv['metadata']['annotations']['containerImage'] = os.getenv('OPERATOR_PULLSPEC', '')
csv['metadata']['annotations']['features.operators.openshift.io/disconnected'] = 'true'
csv['metadata']['annotations']['features.operators.openshift.io/fips-compliant'] = 'true'
csv['metadata']['annotations']['features.operators.openshift.io/proxy-aware'] = 'false'
csv['metadata']['annotations']['features.operators.openshift.io/tls-profiles'] = 'false'
csv['metadata']['annotations']['features.operators.openshift.io/token-auth-aws'] = 'false'
csv['metadata']['annotations']['features.operators.openshift.io/token-auth-azure'] = 'false'
csv['metadata']['annotations']['features.operators.openshift.io/token-auth-gcp'] = 'false'
csv['metadata']['annotations']['features.operators.openshift.io/cnf'] = 'false'
csv['metadata']['annotations']['features.operators.openshift.io/cni'] = 'false'
csv['metadata']['annotations']['features.operators.openshift.io/csi'] = 'false'

# inject bundle creation date in pods annotations
podMeta = csv['spec']['install']['spec']['deployments'][0]['spec']['template']['metadata']
if 'annotations' in podMeta:
   podMeta['annotations']['bundleCreatedAt'] = created_at
else:
   podMeta['annotations'] = {'bundleCreatedAt': created_at}

for env in csv['spec']['install']['spec']['deployments'][0]['spec']['template']['spec']['containers'][0]['env']:
   if env['name'] == 'DOWNSTREAM_DEPLOYMENT':
      env['value'] = "true"
   if env['name'] == 'RELATED_IMAGE_EBPF_AGENT':
      env['value'] = ebpf_image
   if env['name'] == 'RELATED_IMAGE_FLOWLOGS_PIPELINE':
      env['value'] = flp_image
   if env['name'] == 'RELATED_IMAGE_CONSOLE_PLUGIN':
      env['value'] = console_image

for container in csv['spec']['install']['spec']['deployments'][0]['spec']['template']['spec']['containers']:
   if container['name'] == "kube-rbac-proxy":
      container["image"] = "registry.redhat.io/rhacm2/kube-rbac-proxy-rhel8@sha256:c39fd7bf90c92d2062b2e3b264ad146c345debaa8453302c481b4900a058811d"

csv['spec']['install']['spec']['deployments'][0]['spec']['template']['spec']['containers'][0]['image'] = operator_image

# replaces upstream description by something more OpenShift'ish
file = open(desc_file_name,mode='r')
csv['spec']['description'] = file.read()
file.close()

csv['spec']['displayName'] = 'Network Observability'
csv['spec']['maturity'] = 'stable'

# remove relatedImages from spec as it is picked up from ENV instead (having them in both places generates a build error)
csv['spec'].pop('relatedImages', None)

csv['spec']['version'] = version
csv['spec']['replaces'] = 'network-observability-operator.v{}'.format(replaces)

dump_manifest(os.getenv('TARGET_CSV_FILE'), csv)
