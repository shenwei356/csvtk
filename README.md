# csvtk

Another cross-platform, efficient, practical and pretty CSV/TSV toolkit

Yes, you could just use spreadsheet softwares like MS excel to
do most of the job.

Howerver it's all by clicking and typing, which is **not
automatically and time-consuming to repeate**, especially when we want to
apply similar operations with different datasets or purposes.

`csvtk` is **convenient for rapid investigation
and also easy to integrated into analysis pipelines**.
 It could save you much time of writting scripts.

Hope it be helpful for you.


## Features

- **Cross-platform** (Linux/Windows/Mac OS X/OpenBSD/FreeBSD)
- **Light weight and out-of-the-box, no dependencies, no compilation, no configuration**
- **Fast**,  **multiple-CPUs supported**
- **Practical functions supported by N subcommands**
- **Support STDIN and gziped input/output file, easy being used in pipe**
- Most of the subcommands support **unselecting fields** and **fuzzy fields**,
  e.g. `-f "-id,-name"` for all fields except "id" and "name",
  `-F -f "a.*"` for all fields with prefix "a.".


## Subcommands (19 in total)

**Information**

-  `stat` summary of CSV file
-  `stat2` summary of selected number fields

**Format convertion**

-  `pretty` convert CSV to readable aligned table
-  `csv2tab` convert CSV to tabular format
-  `tab2csv` convert tabular format to CSV
-  `space2tab` convert space delimited format to CSV
-  `transpose` transpose CSV data
-  `csv2md` convert CSV to markdown format

**Set operations**

-  `cut` select parts of fields
-  `uniq` unique data without sorting
-  `inter` intersection of multiple files
-  `grep` grep data by selected fields with patterns/regular expressions
-  `filter` filter data by values of selected fields with math expression
-  `join` join multiple CSV files by selected fields

**Edit**

-  `rename` rename column names
-  `rename2` rename column names by regular expression
-  `replace` replace data of selected fields by regular expression
-  `mutate` create new columns from selected fields by regular expression

**Ordering**

-  `sort` sort by selected fields

## Installation

[Download Page](https://github.com/shenwei356/csvtk/releases)

Just [download](https://github.com/shenwei356/csvtk/releases) gzip-compressed
executable file of your operating system, and uncompress it with `gzip -d *.gz` command,
rename it to `csvtk.exe` (Windows) or `csvtk` (other operating systems) for convenience.

You may need to add executable permision by `chmod a+x csvtk`.

You can also add the directory of the executable file to environment variable
`PATH`, so you can run `csvtk` anywhere.

1. For windows, the simplest way is copy it to `C:\WINDOWS\system32`.

2. For Linux, type:

        chmod a+x /PATH/OF/FASTCOV/csvtk
        echo export PATH=\$PATH:/PATH/OF/FASTCOV >> ~/.bashrc

    or simply copy it to `/usr/local/bin`

## Compared to `csvkit`

[csvkit](http://csvkit.readthedocs.org/en/540/)

Features                |  csvtk   |  csvkit   |   Note
:-----------------------|:--------:|:---------:|:---------
Read    Gzip            |   Yes    |  Yes      |
Fields ranges           |   Yes    |  Yes      | e.g. `-f 1-4,6`
**Unselect fileds**     |   Yes    |  --       | e.g. `-1` for excluding first column
**Fuzzy fields**        |   Yes    |  --       | e.g. `ab*` for columns with prefix "ab"
Order-specific fields   |   --     |  Yes      | it means `1,2` is different from `2,1`
**Rename columns**      |   Yes    |  --       | rename with new name(s) or from existed names
Sort by multiple keys   |   Yes    |  Yes      | bash sort like operations
**Sort by number**      |   Yes    |  --       | e.g. `-k 1:n`
**Multiple sort**       |   Yes    |  --       | e.g. `-k 2:r -k 1:nr`
**Pretty output**       |   Yes    |  --       | convert CSV to readable aligned table

Similar tools:

- [csvkit](http://csvkit.readthedocs.org/en/540/) - A suite of utilities for converting to and working with CSV, the king of tabular file formats. http://csvkit.rtfd.org/
- [miller](https://github.com/johnkerl/miller) - Miller is like sed, awk, cut, join, and sort for 
name-indexed data such as CSV and tabular JSON http://johnkerl.org/miller
- [tsv-utils-dlang](https://github.com/eBay/tsv-utils-dlang) - Command line utilities for tab-separated value files written in the D programming language.

## Examples

**Attention**

1. The CSV parser requires all the lines have same number of fields/columns.
 Even lines with spaces will cause error.
2. By default, csvtk think your files have header row, if not, use `-H`.
3. By default, lines starting with `#` will be ignored, if the header row
 starts with `#`, please assign `-C` another rare symbol, e.g. `&`.
4. By default, csvtk handles CSV files, use `-t` for tab-delimited files.

More [examples](http://shenwei356.github.io/csvtk/usage/) and [tutorial](http://shenwei356.github.io/csvtk/tutorial/)

Examples

1. Pretty result

        $ csvtk pretty names.csv
        id   first_name   last_name   username
        11   Rob          Pike        rob
        2    Ken          Thompson    ken
        4    Robert       Griesemer   gri
        1    Robert       Thompson    abc
        NA   Robert       Abel        123

1. Select fields/columns (`cut`)

    - By index: `csvtk cut -f 1,2`
    - By names: `csvtk cut -f first_name,username`
    - **Unselect**: `csvtk cut -f -1,-2` or `csvtk cut -f -first_name`
    - **Fuzzy fields**: `csvtk cut -F -f "*_name,username"`
    - Field ranges: `csvtk cut -f 2-4` for column 2,3,4 or `csvtk cut -f -3--1` for discarding column 1,2,3
    - All fields: `csvtk cut -F -f "*"`

1. Search by selected fields (`grep`) (matched parts will be highlighted as red)

    - By exactly matching: `csvtk grep -f first_name -p Robert -p Rob`
    - By regular expression: `csvtk grep -f first_name -r -p Rob`
    - By pattern list: `csvtk grep -f first_name -P name_list.txt`



1. Rename column names (`rename` and `rename2`)

    - Setting new names: `csvtk rename -f A,B -n a,b` or `csvtk rename -f 1-3 -n a,b,c`
    - Replacing with original names by regular express: `cat ../testdata/c.csv | ./csvtk rename2 -F -f "*" -p "(.*)" -r 'prefix_$1'` for adding prefix to all column names.

1. Edit data with regular expression (`replace`)

    - Remove Chinese charactors:  `csvtk replace -F -f "*_name" -p "\p{Han}+" -r ""`

1. Create new column from selected fields by regular expression (`mutate`)

    - In default, copy a column: `csvtk mutate -f id `
    - Extract prefix of data as group name (get "A" from "A.1" as group name):
      `csvtk mutate -f sample -n group -p "^(.+?)\."`

1. Sort by multiple keys (`sort`)

    - By single column : `csvtk sort -k 1` or `csvtk sort -k last_name`
    - By multiple columns: `csvtk sort -k 1,2` or `csvtk sort -k 1 -k 2` or `csvtk sort -k last_name,age`
    - Sort by number: `csvtk sort -k 1:n` or  `csvtk sort -k 1:nr` for reverse number
    - Complex sort: `csvtk sort -k region -k age:n -k id:nr`

1. Join multiple files by keys (`join`)

    - All files have same key column: `csvtk join -f id file1.csv file2.csv`
    - Files have different key columns: `csvtk join -f "username;username;name" names.csv phone.csv adress.csv -k`

1. Summary of selected number fields: num, sum, min, max, mean, stdev (`stat2`)

        $ cat digitals.tsv
        4       5       6
        1       2       3
        7       8       0
        8       1,000   4

        $ cat digitals.tsv | csvtk stat2 -t -H -f 1-3
        field   num     sum   min     max     mean    stdev
        1         4      20     1       8        5     3.16
        2         4   1,015     2   1,000   253.75   497.51
        3         4      13     0       6     3.25      2.5

## Contact

Email me for any problem when using `csvtk`. shenwei356(at)gmail.com

Or [create an issue](https://github.com/shenwei356/csvtk/issues) to report bugs,
propose new functions or ask for help.

Or [leave a comment](https://shenwei356.github.io/csvtk/usage/#disqus_thread).

## License

[MIT License](https://github.com/shenwei356/csvtk/blob/master/LICENSE)
