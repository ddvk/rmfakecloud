#!/usr/bin/env bash

# URL of extension in Chrome store:
# https://chrome.google.com/webstore/detail/read-on-remarkable/bfhkfdnddlhfippjbflipboognpdpoeh
# Remarkable extension ID:
EXT_ID="bfhkfdnddlhfippjbflipboognpdpoeh"
CHROME_VERSION="31.0.1609.0"
CRX_NAME="remarkable_reader.crx"
declare -a replace_domains
replace_domains=(
  "internal.cloud.remarkable.com"
  "webapp-production-dot-remarkable-production.appspot.com"
)

# Set domain name according to input
MY_DOMAIN="$1" # E.g. 'rmfakecloud.my.domain'. Do not include 'http(s)://' nor a trailing slash.
[ -z "$MY_DOMAIN" ] && echo "Error: provide your domain name as first argument, e.g. 'rmfakecloud.my.domain'" 2>&1 && exit 1
[[ ! "$MY_DOMAIN" =~ ^[^*/?#]*$ ]] && echo "Error: your domain name must not contain any of these characters: */?#" 2>&1 && exit 2
echo "Using domain name: '$MY_DOMAIN'" 2>&1

# Make directories
[ -d 'src' ]       || mkdir src
[ -d 'artefacts' ] || mkdir artefacts

cd src

# Download extension from Chrome store and extract
echo "Downloading official extension..." 2>&1
wget -q -O "$CRX_NAME" "https://clients2.google.com/service/update2/crx?response=redirect&prodversion=$CHROME_VERSION&acceptformat=crx2,crx3&x=id%3D$EXT_ID%26uc"
echo "Unzipping extension..." 2>&1
echo "Any complaint about 'extra bytes at beginning or within zipfile' can probably be ignored..." 2>&1
unzip -q "$CRX_NAME" -d "$EXT_ID/"

cd "$EXT_ID/"
# Substitute reMarkable domains with custom domain
echo "Substituting these domain names for '$MY_DOMAIN':" 2>&1
printf "%s " "${replace_domains[@]}" 2>&1 && echo
regex_match_domain_group="$(printf "%s|" "${replace_domains[@]}" | sed 's/\./\\./g')"
regex_match_domain_group="(${regex_match_domain_group::-1})" # Remove trailing '|' and wrap in parentheses
find ./ -type f -exec sed -E -i "s~(https?://)$regex_match_domain_group~\1$MY_DOMAIN~g" {} \; 

echo "Cleaning up..." 2>&1
cd ..
rm "$CRX_NAME"
cd ..
mv "src/$EXT_ID" "artefacts/$EXT_ID"
echo "Modified extension located at 'artefacts/$EXT_ID'. You can install this into Chrome with 'Load unpacked > ...' in the extensions menu."

