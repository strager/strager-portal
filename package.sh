#!/usr/bin/env bash

out=strager-portal.xpi
rm -f "${out}"
zip "${out}" -- manifest.json

echo "created: ${out}"
