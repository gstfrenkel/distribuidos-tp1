import sys
import re


def parse_file(filename):
    with open(filename, "r") as file:
        content = file.read()

    sections = re.split(r"\n(?=Q\d+:)", content.strip())
    parsed_data = {}

    for section in sections:
        if not section.strip():
            continue

        match = re.match(r"(Q\d+):", section)
        if not match:
            continue

        q_key = match.group(1)
        entries = sorted(
            line.strip() for line in section.splitlines()[1:] if line.strip()
        )
        parsed_data[q_key] = entries

    return parsed_data


def compare_files(file_list):
    parsed_files = {filename: parse_file(filename) for filename in file_list}
    reference_filename = file_list[0]
    reference_data = parsed_files[reference_filename]
    all_files_match = True

    print(f"Using '{reference_filename}' as the reference file for comparison.\n")

    for filename, parsed_data in parsed_files.items():
        if filename == reference_filename:
            continue

        for q_key in sorted(set(reference_data.keys()).union(parsed_data.keys())):
            ref_entries = reference_data.get(q_key, [])
            entries = parsed_data.get(q_key, [])

            if ref_entries != entries:
                all_files_match = False
                print(f"  Difference in section {q_key} of file '{filename}':")

                ref_set, file_set = set(ref_entries), set(entries)
                missing_in_file = sorted(ref_set - file_set)
                extra_in_file = sorted(file_set - ref_set)

                if missing_in_file:
                    print(f"    Missing entries in '{filename}': {missing_in_file}")
                if extra_in_file:
                    print(f"    Extra entries in '{filename}': {extra_in_file}")

    return all_files_match


if __name__ == "__main__":
    if len(sys.argv) < 3:
        print("Usage: python compare_results.py <file1> <file2> [<file3> ...]")
        print("Note: The first file specified (file1) is used as the reference file.")
        sys.exit(1)

    files = sys.argv[1:]
    result = compare_files(files)

    if result:
        print("All files have the same results.")
    else:
        print("Files have differences.")
