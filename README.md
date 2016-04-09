# csvtk

Another cross-platform, efficient and practical CSV/TSV tool kit.

## Features

- **Cross-platform** (Linux/Windows/Mac OS X/OpenBSD/FreeBSD)
- **Light weight and out-of-the-box, no dependencies, no compilation, no configuration**
- **Fast**,  **multiple-CPUs supported**
- **Practical functions supported by N subcommands**
- **Support STDIN and gziped input/output file, easy being used in pipe**

## Installation

Just [download](https://github.com/shenwei356/csvtk/releases) executable file
 of your operating system and rename it to `csvtk.exe` (Windows) or
 `csvtk` (other operating systems) for convenience.
 
You can also add the directory of the executable file to environment variable
`PATH`, so you can run `csvtk` anywhere.

1. For windows, the simplest way is copy it to `C:\WINDOWS\system32`.

2. For Linux, type:

        chmod a+x /PATH/OF/FASTCOV/csvtk
        echo export PATH=\$PATH:/PATH/OF/FASTCOV >> ~/.bashrc

    or simply copy it to `/usr/local/bin`

## Subcommands (16 in total)

**Information**

-  [x] `stat` summary of CSV file

**Format convertion**

-  [x] `csv2tab` convert CSV to tabular format
-  [x] `tab2csv` convert tabular format to CSV
-  [x] `space2tab` convert space delimited format to CSV
-  [x] `transpose` transpose CSV data

**Set operations**

-  [x] `cut` select parts of fields
-  [x] `uniq` unique data without sorting
-  [x] `inter` intersection of multiple files
-  [x] `grep` grep data by selected fields with patterns/regular expressions
-  `filter` filter data by values of selected fields, supporting math/string expression
-  [x] `join` join multiple CSV files by selected fields

**Edit**

-  [x] `rename` rename column names
-  [x] `rename2` rename column names by regular expression
-  [x] `replace` replace data of selected fields by regular expression
-  [x] `mutate` create new columns from selected fields by regular expression

**Ordering**

-  [x] `sort` sort by selected fields

## Compared to `csvkit`

[csvkit](http://csvkit.readthedocs.org/en/540/)

Features                |  csvtk   |  csvkit   |   Note
:-----------------------|:--------:|:---------:|:---------
Read    Gzip            |   Yes    |  Yes      |
Fields ranges           |   Yes    |  Yes      | e.g. `-f 1-4,6`
**Unselect fileds**     |   Yes    |  --       | e.g. `-1` for excluding first column
**Fuzzy fields**        |   Yes    |  --       | e.g. `ab*` for columns with prefix "ab"
Rename columns          |   Yes    |  --       | rename with new name(s) or from existed names
Sort by multiple keys   |   Yes    |  Yes      | bash sort like operations
Sort by number          |   Yes    |  --       | e.g. `-k 1:n`
**Multiple sort**       |   Yes    |  --       | e.g. `-k 2:r -k 1:nr`


to be continued...

## Examples

1. Select fields/columns (`cut`)

    - By index: `csvtk cut -f 1,2`
    - By names: `csvtk cut -f first_name,username`
    - **Unselect**: `csvtk cut -f -1,-2` or `csvtk cut -f -first_name`
    - **Fuzzy fields**: `csvtk cut -F -f "*_name,username"`
    - Field ranges: `csvtk cut -f 2-4` for column 2,3,4 or `csvtk cut -f -3--1` for discarding column 1,2,3
    - All fields: `csvtk cut -F -f "*"`

1. Search by selected fields (`grep`)

    - By exactly matching: `csvtk grep -f first_name -p Robert -p Rob`
    - By regular expression: `csvtk grep -f first_name -r -p Rob`
    - By pattern list: `csvtk grep -f first_name -P name_list.txt`

1. Rename column names (`rename` and `rename2`)

    - Setting new names: `csvtk rename -f A,B -n a,b` or `csvtk rename -f 1-3 -n a,b,c`
    - Replacing with original names by regular express: `cat ../testdata/c.csv | ./csvtk rename2 -F -f "*" -p "(.*)" -r 'prefix_$1'` for adding prefix to all column names.

1. Edit data with regular expression (`replace`)

    - e.g. remove Chinese charactors:  `csvtk replace -F -f "*_name" -p "\p{Han}+" -r ""`

1. Create new column from selected fields by regular expression (`mutate`)

    - In default, copy a column: `csvtk mutate -f id `
    - e.g. extract prefix of data as group name (get "A" from "A.1" as group name):
      `csvtk mutate -f sample -n group -p "^(.+?)\."`

1. Sort by multiple keys (`sort`)

    - By single column : `csvtk sort -k 1` or `csvtk sort -k last_name`
    - By multiple columns: `csvtk sort -k 1,2` or `csvtk sort -k 1 -k 2` or `csvtk sort -k last_name,age`
    - Sort by number: `csvtk sort -k 1:n` or  `csvtk sort -k 1:nr` for reverse number
    - Complex sort: `csvtk sort -k region -k age:n -k id:nr`

1. Join multiple files by keys (`join`)

    - `csvtk join -f "username;username;name" names.csv phone.csv adress.csv`

## Contact

Email me for any problem when using `csvtk`. shenwei356(at)gmail.com

[Create an issue](https://github.com/shenwei356/csvtk/issues) to report bugs,
propose new functions or ask for help.

## License

[MIT License](https://github.com/shenwei356/csvtk/blob/master/LICENSE)
