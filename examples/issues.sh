#!/bin/bash

# colors for fun :P
# source: https://unix.stackexchange.com/a/10065
if test -t 1; then

    # see if it supports colors...
    ncolors=$(tput colors)

    if test -n "$ncolors" && test $ncolors -ge 8; then
        bold="$(tput bold)"
        underline="$(tput smul)"
        standout="$(tput smso)"
        normal="$(tput sgr0)"
        black="$(tput setaf 0)"
        red="$(tput setaf 1)"
        green="$(tput setaf 2)"
        yellow="$(tput setaf 3)"
        blue="$(tput setaf 4)"
        magenta="$(tput setaf 5)"
        cyan="$(tput setaf 6)"
        white="$(tput setaf 7)"
    fi
fi

CACHE_FILE=~/.cache/issues-last
CURSOR=$(cat "${CACHE_FILE}" 2>/dev/null || echo "")
ENDPOINT="https://api.github.com/graphql"
AUTH_HEADER="Authorization=bearer $GITHUB_TOKEN"
ORDER_BY="{field: CREATED_AT, direction: DESC}"
ISSUES_PATH=".viewer.issues"
PAGE_SIZE=10

before() {
	if [ "${1:-x}" != "x" ]; then
		echo "--arg-before $1"
	else
		echo ""
	fi
}

issues() {
	echo "$(gql \
        query viewer issues \
		--header "${AUTH_HEADER}" --endpoint ${ENDPOINT} \
		--arg-last ${PAGE_SIZE} $(before $CURSOR) --arg-orderBy "${ORDER_BY}" \
		"$@")"
}

fetch() {
	echo "$(issues \
		nodes \
		--format "{{ range ${ISSUES_PATH}.nodes }}${bold}#{{ .number }}: ${normal}{{ .title }} ${green}[{{ .url }}]${normal}
{{end}}")"
}

next() {
	echo "$(issues \
		pageInfo startCursor \
		--format "{{ ${ISSUES_PATH}.pageInfo.startCursor }}")"
}

MESSAGE="$(fetch)"
LAST_CURSOR="${CURSOR}"
CURSOR="$(next)"
while [ "${CURSOR}" != "<no value>" ]; do
	MESSAGE="$MESSAGE
$(fetch | sed '1!G;h;$!d')"
	LAST_CURSOR="${CURSOR}"
	CURSOR="$(next)"
done
echo "$LAST_CURSOR" > ${CACHE_FILE}
echo "${MESSAGE}"
