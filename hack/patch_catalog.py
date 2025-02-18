import os
from sys import exit as sys_exit
from ruamel.yaml import YAML
yaml = YAML()
yaml.explicit_start = True

version = os.getenv('VERSION')
bundle_image = os.getenv('BUNDLE_IMAGE_PULLSPEC')
operator_image = os.getenv('OPERATOR_IMAGE_PULLSPEC')
ebpf_image = os.getenv('EBPF_IMAGE_PULLSPEC')
flp_image = os.getenv('FLP_IMAGE_PULLSPEC')
console_image = os.getenv('CONSOLE_IMAGE_PULLSPEC')
package_name = "network-observability-operator"
package_full_name = '{}.v{}'.format(package_name, version)

def load_index(pathn):
   if not pathn.endswith(".yaml"):
      return None
   try:
      with open(pathn, "r") as f:
         return list(yaml.load_all(f))
   except FileNotFoundError:
      print("File can not found")
      exit(6)

def dump_index(pathn, index):
   with open(pathn, "w") as f:
      for o in index:
         yaml.dump(o, f)
   return

index = load_index(os.getenv('TARGET_INDEX_FILE'))

index[0]["image"] = bundle_image

for relatedImage in index[0]["relatedImages"]:
   if relatedImage["image"][0:95] == "registry.redhat.io/network-observability/network-observability-operator-bundle":
      relatedImage["image"] = bundle_image
   elif relatedImage["name"] == "manager":
      relatedImage["image"] = operator_image
   elif relatedImage["name"] == "ebpf_agent":
      relatedImage["image"] = ebpf_image
   elif relatedImage["name"] == "flowlogs_pipeline":
      relatedImage["image"] = flp_image
   elif relatedImage["name"] == "console_plugin":
      relatedImage["image"] = console_image

dump_index(os.getenv('TARGET_INDEX_FILE'), index)
