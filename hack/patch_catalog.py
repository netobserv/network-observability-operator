import os
from ruamel.yaml import YAML
yaml = YAML()
yaml.explicit_start = True

bundle_image = os.getenv('BUNDLE_IMAGE_PULLSPEC')
operator_image = os.getenv('OPERATOR_IMAGE_PULLSPEC')
ebpf_image = os.getenv('EBPF_IMAGE_PULLSPEC')
flp_image = os.getenv('FLP_IMAGE_PULLSPEC')
console_image = os.getenv('CONSOLE_IMAGE_PULLSPEC')

def load_bundle(pathn):
   if not pathn.endswith(".yaml"):
      return None
   try:
      with open(pathn, "r") as f:
         return list(yaml.load_all(f))
   except FileNotFoundError:
      print("File can not found")
      exit(6)

def dump_bundle(pathn, index):
   with open(pathn, "w") as f:
      for o in index:
         yaml.dump(o, f)
   return

bundle = load_bundle(os.getenv('NEW_BUNDLE_FILE'))

bundle[0]["image"] = bundle_image

for relatedImage in bundle[0]["relatedImages"]:
   if relatedImage["name"] == "bundle":
      relatedImage["image"] = bundle_image
   elif relatedImage["name"] == "manager":
      relatedImage["image"] = operator_image
   elif relatedImage["name"] == "ebpf_agent":
      relatedImage["image"] = ebpf_image
   elif relatedImage["name"] == "flowlogs_pipeline":
      relatedImage["image"] = flp_image
   elif relatedImage["name"] == "console_plugin":
      relatedImage["image"] = console_image

dump_bundle(os.getenv('NEW_BUNDLE_FILE'), bundle)
