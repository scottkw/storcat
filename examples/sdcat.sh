#!/bin/bash

echo ""
echo "Tree Command SD Cataloger"
echo "Written by Ken Scott"
echo "Wednesday June 1, 2022"
echo ""

function usage() {

    echo ""
    echo "Usage: sdcat.sh COMMAND ARGS"
    echo "          search | s TERM"
    echo "              TERM  = the term you would like to find in the catalogs"
    echo "          create | c TITLE PATH ROOT"
    echo "              TITLE = title of the catalog as displayed in the HTML catalog"
    echo "              PATH  = the path of the directory you would like cataloged"
    echo "              ROOT  = the root (i.e. root.ext) of the catalog file"
    echo ""

}

function createCatalog() {

    # Create JSON catalog file
    echo "Creating JSON catalog file..."

    tree -h --du -f -J $2 -o $2/$3-temp.json

    # Remove trailing commas from JSON output (specifically for tree versions < 1.8)
    perl -MJSON -e '@text=(<>);print to_json(from_json("@text", {relaxed=>1}))' $2/$3-temp.json > $2/$3.json
    rm $2/$3-temp.json

    echo "JSON catalog file complete."

    # Create HTML catalog file
    echo "Creating HTML catalog file..."

    tree -h --du -f -T "$1" --nolinks -H $2 -o $2/$3.html

    echo "HTML catalog file complete."
    echo ""

}

function searchCatalogs() {

    # Loop through all the .json files
    for FILE in *.json; do 
        
        # Strip .json filename down to the base name
        catalog_name=$(basename "$FILE" .json)

        # Uppercase the base name
        catalog_name=${catalog_name^^}

        # Search for the given search term with the uppercase base name as a prefix
        jq -r '.. | .name?' $FILE | grep -i "$1" | awk '$0="'$catalog_name': "$0""'

    done

}

# Check for COMMAND argument
[[ -z "$1" ]] && { echo "Missing command."; echo ""; usage; exit 1; }

if [ "${1^^}" = "SEARCH" ] || [ "${1^^}" = "S" ]; then
    [[ -z "$2" ]] && { echo "Missing search term."; echo ""; usage; exit 1; }
    searchCatalogs "$2"
    echo ""
    echo "---===<<< SEARCH COMPLETE >>>===---"
elif [ "${1^^}" = "CREATE" ] || [ "${1^^}" = "C" ]; then
    [[ -z "$2" ]] && { echo "Missing argument."; echo ""; usage; exit 1; }
    [[ -z "$3" ]] && { echo "Missing argument."; echo ""; usage; exit 1; }
    [[ -z "$4" ]] && { echo "Missing argument."; echo ""; usage; exit 1; }
    createCatalog "$2" "$3" "$4"
    echo ""
    echo "---===<<< CATALOG CREATION COMPLETE >>>===---"
else
    { echo "Invalid command."; echo ""; usage; exit 1; }
fi

echo ""


