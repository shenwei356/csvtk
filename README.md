# csvtk - another cross-platform, efficient, practical and pretty CSV/TSV toolkit

**Documents:** [http://bioinf.shenwei.me/csvtk](http://bioinf.shenwei.me/csvtk/) 
( [**Usage**](http://bioinf.shenwei.me/csvtk/usage/)  and [**Tutorial**](http://bioinf.shenwei.me/csvtk/tutorial/))

**Source code:**  [https://github.com/shenwei356/csvtk](https://github.com/shenwei356/csvtk) [![GitHub stars](https://img.shields.io/github/stars/shenwei356/csvtk.svg?style=social&label=Star&?maxAge=2592000)](https://github.com/shenwei356/csvtk)
[![license](https://img.shields.io/github/license/shenwei356/csvtk.svg?maxAge=2592000)](https://github.com/shenwei356/csvtk/blob/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/shenwei356/csvtk)](https://goreportcard.com/report/github.com/shenwei356/csvtk)


**Latest version:** [![Latest Stable Version](https://img.shields.io/github/release/shenwei356/csvtk.svg?style=flat)](https://github.com/shenwei356/csvtk/releases)
[![Github Releases](https://img.shields.io/github/downloads/shenwei356/csvtk/latest/total.svg?maxAge=3600)](http://bioinf.shenwei.me/csvtk/download/)

`csvtk` is **convenient for rapid investigation
and also easy to integrated into analysis pipelines**.
 It could save you much time of writing scripts.

## Features

- **Cross-platform** (Linux/Windows/Mac OS X/OpenBSD/FreeBSD)
- **Light weight and out-of-the-box, no dependencies, no compilation, no configuration**
- **Fast**,  **multiple-CPUs supported**
- **Practical functions supported by N subcommands**
- **Support STDIN and gziped input/output file, easy being used in pipe**
- Most of the subcommands support ***unselecting fields*** and ***fuzzy fields***,
  e.g. `-f "-id,-name"` for all fields except "id" and "name",
  `-F -f "a.*"` for all fields with prefix "a.".
- **Support common plots** (see [usage](http://bioinf.shenwei.me/csvtk/usage/#plot))


## Subcommands

20 in total.

**Information**

-  `stat` summary of CSV file
-  `stat2` summary of selected number fields

**Format conversion**

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

**Ploting**

- `plot` see [usage](http://bioinf.shenwei.me/csvtk/usage/#plot)
    - `plot hist` histogram
    - `plot box` boxplot
    - `plot line` line plot and scatter plot

## Installation

[Download Page](https://github.com/shenwei356/csvtk/releases)

`csvtk` is implemented in [Go](https://golang.org/) programming language,
 executable binary files **for most popular operating systems** are freely available
  in [release](https://github.com/shenwei356/csvtk/releases) page.

Just [download](https://github.com/shenwei356/csvtk/releases) compressed
executable file of your operating system, and uncompress it with `tar -zxvf *.tar.gz` command.

You can add the directory of the executable file to environment variable
`PATH`, so you can run `csvtk` anywhere.

1. For Unix-like systems
    1. If you have root privilege simply copy it to `/usr/local/bin`:

            sudo cp csvtk /usr/local/bin/

    1. Or add the current directory of the executable file to environment variable
    `PATH`:

            echo export PATH=\$PATH:\"$(pwd)\" >> ~/.bashrc
            source ~/.bashrc

1. **For windows**, just copy `csvtk.exe` to `C:\WINDOWS\system32`.

For Go developer, just one command:

    go get -u github.com/shenwei356/csvtk/csvtk

## Compared to `csvkit`

[csvkit](http://csvkit.readthedocs.org/en/540/)

Features                |  csvtk   |  csvkit   |   Note
:-----------------------|:--------:|:---------:|:---------
Read    Gzip            |   Yes    |  Yes      |
Fields ranges           |   Yes    |  Yes      | e.g. `-f 1-4,6`
**Unselect fileds**     |   Yes    |  --       | e.g. `-1` for excluding first column
**Fuzzy fields**        |   Yes    |  --       | e.g. `ab*` for columns with name prefix "ab"
Order-specific fields   |   Yes    |  Yes      | it means `-f 1,2` is different from `-f 2,1`
**Rename columns**      |   Yes    |  --       | rename with new name(s) or from existed names
Sort by multiple keys   |   Yes    |  Yes      | bash sort like operations
**Sort by number**      |   Yes    |  --       | e.g. `-k 1:n`
**Multiple sort**       |   Yes    |  --       | e.g. `-k 2:r -k 1:nr`
**Pretty output**       |   Yes    |  --       | convert CSV to readable aligned table

Similar tools:

- [csvkit](http://csvkit.readthedocs.org/en/540/) - A suite of utilities for converting to and working with CSV, the king of tabular file formats. http://csvkit.rtfd.org/
- [xsv](https://github.com/BurntSushi/xsv) - A fast CSV toolkit written in Rust.
- [miller](https://github.com/johnkerl/miller) - Miller is like sed, awk, cut, join, and sort for
name-indexed data such as CSV and tabular JSON http://johnkerl.org/miller
- [tsv-utils-dlang](https://github.com/eBay/tsv-utils-dlang) - Command line utilities for tab-separated value files written in the D programming language.

## Examples

**Attention**

1. The CSV parser requires all the lines have same number of fields/columns.
    Even lines with spaces will cause error.
2. By default, csvtk thinks your files have header row, if not, use "-H".
3. By default, lines starting with '#' will be ignored, if the header row
    starts with '#', please assign "-C" another rare symbol, e.g. '$'.
4. By default, csvtk handles CSV files, use "-t" for tab-delimited files.
5. If " exists in tab-delimited files, use "-l".

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
    - Remore rows containing missing data (NA): `csvtk grep -F -f "*" -r -p "^$" -v `


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

1. Filter by numbers (`filter`)

    - single field: `csvtk filter -f "id>0"`
    - multiple fields: `csvtk filter -f "1-3>0"`
    - using `--any` to print record if any of the field satisfy the condition: `csvtk filter -f "1-3>0" --any`
    - fuzzy fields: `csvtk filter -F -f "A*!=0"`

1. ploting
    - plot histogram with data of the second column:
     `csvtk -t plot hist testdata/grouped_data.tsv.gz -f 2`
    ![histogram.png](testdata/figures/histogram.png)
    - plot boxplot with data of the "GC Content" (third) column,
    group information is the "Group" column.
    `csvtk -t plot box testdata/grouped_data.tsv.gz -g "Group" -f "GC Content"  --width 3`
    ![boxplot.png](testdata/figures/boxplot.png)
    -  plot horiz boxplot with data of the "Length" (second) column,
    group information is the "Group" column.
    `csvtk -t plot box testdata/grouped_data.tsv.gz -g "Group" -f "Length"  --height 3 --width 5 --horiz --title "Horiz box plot"`
    ![boxplot2.png](testdata/figures/boxplot2.png)
    - plot line plot with X-Y data
    `csvtk -t plot line testdata/xy.tsv -x X -y Y -g Group`
    ![lineplot.png](testdata/figures/lineplot.png)
    - plot scatter plot with X-Y data
    `csvtk -t plot line testdata/xy.tsv -x X -y Y -g Group --scatter `
    ![scatter.png](testdata/figures/scatter.png)

## Contact

Email me for any problem when using `csvtk`. shenwei356(at)gmail.com

Or [create an issue](https://github.com/shenwei356/csvtk/issues) to report bugs,
propose new functions or ask for help.

Or [leave a comment](https://shenwei356.github.io/csvtk/usage/#disqus_thread).

## License

[MIT License](https://github.com/shenwei356/csvtk/blob/master/LICENSE)
