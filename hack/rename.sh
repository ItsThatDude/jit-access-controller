#!/usr/bin/env bash
# Recursively rename files and directories replacing JIT-prefixed names
# (e.g., JitAccessRequest, JITAccessGrant) with non-JIT versions,
# preserving the original case pattern.

set -e

# Map of JIT-prefixed tokens to their replacements (without JIT)
declare -A replacements=(
  ["JitAccessRequest"]="AccessRequest"
  ["JitAccessGrant"]="AccessGrant"
  ["JitAccessPolicy"]="AccessPolicy"
  ["JitAccessPolicies"]="AccessPolicies"
  ["JitAccessResponses"]="AccessResponses"
  ["JitAccessResponse"]="AccessResponse"
)

# Process all files and directories recursively (deepest first)
find . -depth | while IFS= read -r file; do
   # Skip the current directory itself
  [[ "$file" == "." ]] && continue
  [[ -e "$file" ]] || continue

  dirname=$(dirname "$file")
  basename=$(basename "$file")
  newbase="$basename"

  for key in "${!replacements[@]}"; do
    repl="${replacements[$key]}"

    # Handle PascalCase (JitAccessRequest -> AccessRequest)
    newbase="${newbase//${key}/${repl}}"

    # Handle all-uppercase (JITACCESSREQUEST -> ACCESSREQUEST)
    key_upper=$(echo "$key" | tr '[:lower:]' '[:upper:]')
    repl_upper=$(echo "$repl" | tr '[:lower:]' '[:upper:]')
    newbase="${newbase//${key_upper}/${repl_upper}}"

    # Handle all-lowercase (jitaccessrequest -> accessrequest)
    key_lower=$(echo "$key" | tr '[:upper:]' '[:lower:]')
    repl_lower=$(echo "$repl" | tr '[:upper:]' '[:lower:]')
    newbase="${newbase//${key_lower}/${repl_lower}}"

    # Handle camelCase (jitAccessRequest -> accessRequest)
    key_camel="$(tr '[:upper:]' '[:lower:]' <<<"${key:0:1}")${key:1}"
    repl_camel="$(tr '[:upper:]' '[:lower:]' <<<"${repl:0:1}")${repl:1}"
    newbase="${newbase//${key_camel}/${repl_camel}}"
  done

  newpath="$dirname/$newbase"
  if [[ "$file" != "$newpath" ]]; then
    echo "Renaming: $file → $newpath"
    mv "$file" "$newpath"
  fi
done
