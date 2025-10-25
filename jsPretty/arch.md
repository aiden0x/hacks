# Beautify JS Files

## Main Function

1. Create output directory:
    - output directory name -> "beautified".
    - create the directory with `os.MkdirAll` with `0755` permissions and errors.
2. Handle Input: single file from CLI or list from stdin nand handle errors (scanner error, empty input error).
3. loop in input files and // call `Process File function`

## Functions

1. `Process File Function` (takes a path and output directory in type string):
    - check for `.js` files and handle errors.
    - get the file info using `os.Stat` and check if it's a regular file then handle errors.
    - Beautify the file:
        - using `jsbeautifier` beatify the file with it's path using default options and handle errors.
    - Write to output directory:
        - using `path/filepath` module join the output path using the output directory and file path base.
        - using `os.WriteFile` write the file using output path and beautified array with the file info mode perm and handle errors.
    - Results:
        - print the total results `[+] Beautified path -> output path`
