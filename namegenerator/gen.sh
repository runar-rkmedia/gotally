#!/usr/bin/env bash
#
gen() {
  if [ ! -f "$2" ]; then
    echo "$2 does not exist."
    return
  fi


# Writes the file into a go-compatible list of strings.
cat << EOF > $1
package namegenerator

$4
var $3 = []string{
$(cat $2 | sed 's/.*/\u&/' | sed 's/\(.*\)/"\1",/' | sort | uniq)
}

EOF

gofmt -w  ./namegenerator/subjectives.go

echo "Wrote '$1' based on '$1'. Linecount: $(wc -l $2 | cut -d' ' -f1)"
}

gen ./namegenerator/subjectives.go ./namegenerator/subjectives subjectives "// generated list of subjectives. For this application, it is list of mathematical terms."
gen ./namegenerator/superlatives.go ./namegenerator/superlatives superlatives "// generated list of superlatives."
gen ./namegenerator/adjectives.go ./namegenerator/adjectives adjectives "// generated list of adjectives."
