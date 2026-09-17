import glob
import os

this_branch = "test_output"
main_branch = "test_output_main"

for file in glob.glob(f"{this_branch}/**/*"):
    reference = file.replace(this_branch, main_branch)
    
    # if this is a md file with actual text content, compare the text
    if (file.endswith(".md") or file.endswith(".svg")) and os.path.exists(reference):
        with open(file, "r") as f1, open(reference, "r") as f2:
            if f1.read().strip() != f2.read().strip():
                print(f"{file} and {reference} are different")

print("All tests completed")