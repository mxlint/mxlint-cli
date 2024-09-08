#! /usr/bin/env python3

import json
import os
import sys
import argparse
import yaml

"""
# METADATA
# scope: package
# title: No more than 15 persistent entities within one domain model
# description: The bigger the domain models, the harder they will be to maintain. It adds complexity to your security model as well. The smaller the modules, the easier to reuse.
# authors:
# - Xiwen Cheng <x@cinaq.com>
# custom:
#  category: Maintainability
#  rulename: NumberOfEntities
#  severity: MEDIUM
#  rulenumber: 002_0001
#  remediation: Split domain model into multiple modules.
#  input: "*/DomainModels$DomainModel.yaml"
"""

def parse(policy_file: str):
    """
    Parse the policy.
    """
    metadata_lines = []

    with open(policy_file, "r") as f:
        for line in f.readlines():
            if line.startswith("# METADATA"):
                continue
            clean_line = line.replace("# ", "", 1).strip()
            if line.startswith("#"):
                metadata_lines.append(clean_line)
            else:
                break
    raw_yaml = "\n".join(metadata_lines)
    metadata = yaml.load(raw_yaml, Loader=yaml.SafeLoader)
    return metadata

def generate_policies_docs(policies_dir: str, output_dir: str):
    """
    Generate the policies documentation from the policies directory.
    """
    for policy in os.listdir(policies_dir):
        out_dir = os.path.join(output_dir, policy)
        in_dir = os.path.join(policies_dir, policy)
        os.makedirs(out_dir, exist_ok=True)

        if os.path.isdir(in_dir):
            for file in os.listdir(in_dir):
                if file.endswith(".rego") and not file.endswith("_test.rego"):
                    in_file = os.path.join(in_dir, file)
                    out_file = os.path.join(out_dir, file.replace(".rego", ".md"))
                    generate_policy_docs(in_file, out_file)

def get_test_file(policy_file: str):
    """
    Get the test file for the policy.
    """
    test_file = policy_file.replace(".rego", "_test.rego")
    if not os.path.exists(test_file):
        return "# No test file found"

    with open(test_file, "r") as f:
        body = f.read()
    return body

def generate_policy_docs(policy_file: str, output_file: str):
    policy = parse(policy_file)

    title = policy.get("title", "Untitled")
    description = policy.get("description", "No description provided")
    remediation = policy.get("remediation", "No remediation provided")
    del policy["title"]
    del policy["description"]
    del policy["remediation"]
    del policy["custom"]

    with open(output_file, "w") as f:
        f.write(f"# {title}\n")
        f.write(f"{remediation}\n")
        f.write(f"## Metadata\n")
        f.write(f"## Description\n")
        f.write(f"{description}\n")
        f.write(f"## Remediation\n")
        f.write(f"```yaml\n")
        f.write(f"{yaml.dump(policy)}\n")
        f.write(f"```\n")
        f.write(f"## Test cases\n")
        f.write(f"```rego\n")
        f.write(f"{get_test_file(policy_file)}\n")
        f.write(f"```\n")


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("-p", "--policies", type=str, default="policies", help="The directory containing the policy definitions")
    parser.add_argument("-o", "--output", type=str, default="docs/policies", help="The directory to output the policies documentation")
    args = parser.parse_args()

    generate_policies_docs(args.policies, args.output)

