import jsonref
import json
from os.path import dirname

if __name__ == "__main__":

    base_path = dirname(__file__)
    base_uri = "file://{}/".format(base_path)

    with open("./tasks.json", "r") as data_file:
        data = jsonref.load(data_file, base_uri=base_uri,
                            jsonschema=True, merge_props=True)


    with open("../tasks.json", "w") as o:
        out = json.dumps(data, indent=2)
        o.write(out)
        o.write("\n")
