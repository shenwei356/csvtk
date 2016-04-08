# csvtk

Another cross-platform, efficient and practical CSV/TSV tool kit.

## Features

- **Cross-platform** (Linux/Windows/Mac OS X/OpenBSD/FreeBSD)
- **Light weight and out-of-the-box, no dependencies, no compilation, no configuration**
- **Fast**,  **multiple-CPUs supported**
- **Practical functions supported by N subcommands**
- **Support STDIN and gziped input/output file, easy being used in pipe**

## Download

Try the preview version [v0.1](https://github.com/shenwei356/csvtk/releases/tag/v0.1).

## Subcommands

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
-  `join` join multiple CSV files by selected fields
-  `split` split data to multiple files by values of selected fields

**Edit**

-  `replace` replace data of selected fields by regular expression
-  `mutate` create new columns from selected fields by regular expression

**Ordering**

-  `sort` sort by selected fields

## Compared to `csvkit`

[csvkit](http://csvkit.readthedocs.org/en/540/)

Features                |  csvtk   |  csvkit
:-----------------------|:--------:|:--------:
Read    Gzip            |   Yes    |  Yes
**Unselect fileds**     |   Yes    |  No
**Fuzzy fields**        |   Yes    |  No

to be continued...

## Examples

1. Select fields/columns

    1. By index: `csvtk cut -f 1,2`
    1. By names: `csvtk cut -f first_name,username`
    1. **Unselect**: `csvtk cut -f -1,-2` or `csvtk cut -f -first_name`
    1. **Fuzzy fields**: `csvtk cut -F -f "*_name,username"`

1. Grep by selected fields

    1. By exactly matching: `csvtk grep -f first_name -p Robert -p Rob`
    1. By regular expression: `csvtk grep -f first_name -r -p Rob`
    1. By pattern list: `csvtk grep -f first_name -P name_list.txt`

## Contact

Email me for any problem when using `csvtk`. shenwei356(at)gmail.com

[Create an issue](https://github.com/shenwei356/csvtk/issues) to report bugs,
propose new functions or ask for help.

## License

[MIT License](https://github.com/shenwei356/csvtk/blob/master/LICENSE)
