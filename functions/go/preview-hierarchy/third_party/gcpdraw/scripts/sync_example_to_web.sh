#!/bin/bash
set -eu

for file in examples/*; do
    file_name="${file%.*}"
    file_extension="${file##*.}"
    js_file=${file_name}.js

    if [ "$file_extension" != "txt" ]
    then
      echo "skip ${file}"
      continue
    fi

    dst_path=web/client/src/config/examples/${js_file}
    echo "sync ${file} to ${dst_path}"

    # delete current example
    rm -f "${dst_path}"

    # create new example
    touch "${dst_path}"
    {
      echo -n 'const code = `'
      cat examples/"${file}"
      echo '`'
      echo 'export default code;'
    } >> "${dst_path}"

done
