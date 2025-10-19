#!/bin/bash

set -euo pipefail

SCHEMA_FILE="$1"
DB_FILE="$2"

if [[ -z "$SCHEMA_FILE" || -z "$DB_FILE" ]]; then
    echo "usage: $0 schema_file db_file"
    exit 1
fi

SQL_FILE=$(mktemp)

DROP_ORDER=(
    "index triggers"
    "index definitions"
    "view definitions"
)

CREATE_ORDER=(
    "view definitions"
    "index definitions"
    "index triggers"
    "index population"
)

(
    echo "pragma foreign_keys=off;"
    echo "begin transaction;"

    for SECTION in "${DROP_ORDER[@]}"; do
        CONTENT=$(awk "/-- begin: $SECTION/,/-- end: $SECTION/" "$SCHEMA_FILE")
        case "$SECTION" in
            "index triggers")
                NAMES=$(echo "$CONTENT" | awk '/^create trigger / {print $3}')
                for NAME in $NAMES; do
                    echo "drop trigger if exists \"$NAME\";"
                done
                ;;
            "index definitions")
                NAMES=$(echo "$CONTENT" | awk '/^create virtual table / {print $4}')
                for NAME in $NAMES; do
                    echo "drop table if exists \"$NAME\";"
                done
                ;;
            "view definitions")
                NAMES=$(echo "$CONTENT" | awk '/^create view / {print $3}')
                for NAME in $NAMES; do
                    echo "drop view if exists \"$NAME\";"
                done
                ;;
        esac
    done

    for SECTION in "${CREATE_ORDER[@]}"; do
        awk "/-- begin: $SECTION/,/-- end: $SECTION/" "$SCHEMA_FILE"
        echo ""
    done

    echo "commit;"
    echo "pragma foreign_keys=on;"
    echo "vacuum;"
) > "$SQL_FILE"

echo "Executing rebuild..."
if ! sqlite3 -echo "$DB_FILE" < "$SQL_FILE"; then
    echo "Error: rebuild failed."
    rm "$SQL_FILE"
    exit 1
fi

rm "$SQL_FILE"
echo "Rebuild complete."
