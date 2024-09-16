import os
from sys import exit as sys_exit
from datetime import datetime
from ruamel.yaml import YAML
yaml = YAML()
yaml.explicit_start = True

version = os.getenv('VERSION')
bundle_image = os.getenv('BUNDLE_IMAGE_PULLSPEC')
package_name = "netobserv-operator"
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

index[1]["image"] = bundle_image

# Changing package name
index[1]["name"] = package_full_name
index[2]["entries"][0]["name"] = package_full_name
for prop in index[1]["properties"]:
   if prop["type"] == "olm.package":
      prop["value"]["version"] = version

# Setting channel to stable
index[0]["defaultChannel"] = "stable"
index[2]["name"] = "stable"
dump_index(os.getenv('TARGET_INDEX_FILE'), index)
