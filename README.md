# csvtk - A cross-platform, efficient and practical CSV/TSV toolkit

- **Documents:** [http://bioinf.shenwei.me/csvtk](http://bioinf.shenwei.me/csvtk/)
( [**Usage**](http://bioinf.shenwei.me/csvtk/usage/)  and [**Tutorial**](http://bioinf.shenwei.me/csvtk/tutorial/)). [中文介绍](http://bioinf.shenwei.me/csvtk/chinese)
- **Source code:**  [https://github.com/shenwei356/csvtk](https://github.com/shenwei356/csvtk) [![GitHub stars](https://img.shields.io/github/stars/shenwei356/csvtk.svg?style=social&label=Star&?maxAge=2592000)](https://github.com/shenwei356/csvtk)
[![license](https://img.shields.io/github/license/shenwei356/csvtk.svg?maxAge=2592000)](https://github.com/shenwei356/csvtk/blob/master/LICENSE)
[![Build Status](https://travis-ci.org/shenwei356/csvtk.svg?branch=master)](https://travis-ci.org/shenwei356/csvtk)
- **Latest version:** [![Latest Stable Version](https://img.shields.io/github/release/shenwei356/csvtk.svg?style=flat)](https://github.com/shenwei356/csvtk/releases)
[![Github Releases](https://img.shields.io/github/downloads/shenwei356/csvtk/latest/total.svg?maxAge=3600)](http://bioinf.shenwei.me/csvtk/download/)
[![Cross-platform](https://img.shields.io/badge/platform-any-ec2eb4.svg?style=flat)](http://bioinf.shenwei.me/csvtk/download/)
[![Anaconda Cloud](https://anaconda.org/bioconda/csvtk/badges/version.svg)](https://anaconda.org/bioconda/csvtk)


## Introduction

Similar to FASTA/Q format in field of Bioinformatics,
CSV/TSV formats are basic and ubiquitous file formats in both Bioinformatics and data science.

People usually use spreadsheet software like MS Excel to process table data.
However this is all by clicking and typing, which is **not
automated and is time-consuming to repeat**, especially when you want to
apply similar operations with different datasets or purposes.

***You can also accomplish some CSV/TSV manipulations using shell commands,
but more code is needed to handle the header line. Shell commands do not
support selecting columns with column names either.***

`csvtk` is **convenient for rapid data investigation
and also easy to integrate into analysis pipelines**.
It could save you lots of time in (not) writing Python/R scripts.


## Table of Contents

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Features](#features)
- [Subcommands](#subcommands)
- [Installation](#installation)
- [Bash-completion](#bash-completion)
- [Compared to `csvkit`](#compared-to-csvkit)
- [Examples](#examples)
- [Acknowledgements](#acknowledgements)
- [Contact](#contact)
- [License](#license)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->


## Features

- **Cross-platform** (Linux/Windows/Mac OS X/OpenBSD/FreeBSD)
- **Light weight and out-of-the-box, no dependencies, no compilation, no configuration**
- **Fast**,  **multiple-CPUs supported** (some commands)
- **Practical functions provided by N subcommands**
- **Support STDIN and gziped input/output file, easy being used in pipe**
- Most of the subcommands support ***unselecting fields*** and ***fuzzy fields***,
  e.g. `-f "-id,-name"` for all fields except "id" and "name",
  `-F -f "a.*"` for all fields with prefix "a.".
- **Support some common plots** (see [usage](http://bioinf.shenwei.me/csvtk/usage/#plot))
- <del>Seamlessly support for data with meta line (e.g., `sep=,`) of separator declaration used by MS Excel</del>

## Subcommands

39 subcommands in total.

**Information**

- [`headers`](https://bioinf.shenwei.me/csvtk/usage/#headers): prints headers
- [`dim`](https://bioinf.shenwei.me/csvtk/usage/#dim): dimensions of CSV file
- [`summary`](https://bioinf.shenwei.me/csvtk/usage/#summary): summary statistics of selected digital fields (groupby group fields)
- [`watch`](https://bioinf.shenwei.me/csvtk/usage/#watch): online monitoring and histogram of selected field
- [`corr`](https://bioinf.shenwei.me/csvtk/usage/#corr): calculate Pearson correlation between numeric columns

**Format conversion**

- [`pretty`](https://bioinf.shenwei.me/csvtk/usage/#pretty): converts CSV to readable aligned table
- [`csv2tab`](https://bioinf.shenwei.me/csvtk/usage/#csv2tab): converts CSV to tabular format
- [`tab2csv`](https://bioinf.shenwei.me/csvtk/usage/#tab2csv): converts tabular format to CSV
- [`space2tab`](https://bioinf.shenwei.me/csvtk/usage/#space2tab): converts space delimited format to CSV
- [`transpose`](https://bioinf.shenwei.me/csvtk/usage/#transpose): transposes CSV data
- [`csv2md`](https://bioinf.shenwei.me/csvtk/usage/#csv2md): converts CSV to markdown format
- [`csv2json`](https://bioinf.shenwei.me/csvtk/usage/#csv2json): converts CSV to JSON format
- [`xlsx2csv`](https://bioinf.shenwei.me/csvtk/usage/#xlsx2csv): converts XLSX to CSV format

**Set operations**

- [`head`](https://bioinf.shenwei.me/csvtk/usage/#head): prints first N records
- [`concat`](https://bioinf.shenwei.me/csvtk/usage/#concat): concatenates CSV/TSV files by rows
- [`sample`](https://bioinf.shenwei.me/csvtk/usage/#sample): sampling by proportion
- [`cut`](https://bioinf.shenwei.me/csvtk/usage/#cut): selects parts of fields
- [`grep`](https://bioinf.shenwei.me/csvtk/usage/#grep): greps data by selected fields with patterns/regular expressions
- [`uniq`](https://bioinf.shenwei.me/csvtk/usage/#uniq): unique data without sorting
- [`freq`](https://bioinf.shenwei.me/csvtk/usage/#freq): frequencies of selected fields
- [`inter`](https://bioinf.shenwei.me/csvtk/usage/#inter): intersection of multiple files
- [`filter`](https://bioinf.shenwei.me/csvtk/usage/#filter): filters rows by values of selected fields with arithmetic expression
- [`filter2`](https://bioinf.shenwei.me/csvtk/usage/#filter2): filters rows by awk-like arithmetic/string expressions
- [`join`](https://bioinf.shenwei.me/csvtk/usage/#join): join files by selected fields (inner, left and outer join)
- [`split`](https://bioinf.shenwei.me/csvtk/usage/#split) splits CSV/TSV into multiple files according to column values
- [`splitxlsx`](https://bioinf.shenwei.me/csvtk/usage/#splitxlsx): splits XLSX sheet into multiple sheets according to column values
- [`collapse`](https://bioinf.shenwei.me/csvtk/usage/#collapse): collapses one field with selected fields as keys
- [`comb`](https://bioinf.shenwei.me/csvtk/usage/#comb): compute combinations of items at every row

**Edit**

- [`add-header`](https://bioinf.shenwei.me/csvtk/usage/#add-header): add column names
- [`del-header`](https://bioinf.shenwei.me/csvtk/usage/#del-header): delete column names
- [`rename`](https://bioinf.shenwei.me/csvtk/usage/#rename): renames column names with new names
- [`rename2`](https://bioinf.shenwei.me/csvtk/usage/#rename2): renames column names by regular expression
- [`replace`](https://bioinf.shenwei.me/csvtk/usage/#replace): replaces data of selected fields by regular expression
- [`mutate`](https://bioinf.shenwei.me/csvtk/usage/#mutate): creates new columns from selected fields by regular expression
- [`mutate2`](https://bioinf.shenwei.me/csvtk/usage/#mutate2): creates new column from selected fields by awk-like arithmetic/string expressions
- [`sep`](https://bioinf.shenwei.me/csvtk/usage/#sep): separate column into multiple columns
- [`gather`](https://bioinf.shenwei.me/csvtk/usage/#gather): gathers columns into key-value pairs

**Ordering**

- [`sort`](https://bioinf.shenwei.me/csvtk/usage/#sort): sorts by selected fields

**Ploting**

- [`plot`](https://bioinf.shenwei.me/csvtk/usage/#plot) see [usage](http://bioinf.shenwei.me/csvtk/usage/#plot)
    - [`plot hist`](https://bioinf.shenwei.me/csvtk/usage/#hist) histogram
    - [`plot box`](https://bioinf.shenwei.me/csvtk/usage/#box) boxplot
    - [`plot line`](https://bioinf.shenwei.me/csvtk/usage/#line) line plot and scatter plot

**Misc**

- [`cat`](https://bioinf.shenwei.me/csvtk/usage/#cat) stream file and report progress
- [`version`](https://bioinf.shenwei.me/csvtk/usage/#version)   print version information and check for update
- [`genautocomplete`](https://bioinf.shenwei.me/csvtk/usage/#genautocomplete) generate shell autocompletion script


## Installation

[Download Page](https://github.com/shenwei356/csvtk/releases)

`csvtk` is implemented in [Go](https://golang.org/) programming language,
 executable binary files **for most popular operating systems** are freely available
  in [release](https://github.com/shenwei356/csvtk/releases) page.

#### Method 1: Download binaries (latest stable/dev version)

Just [download](https://github.com/shenwei356/csvtk/releases) compressed
executable file of your operating system,
and decompress it with `tar -zxvf *.tar.gz` command or other tools.
And then:

1. **For Linux-like systems**
    1. If you have root privilege simply copy it to `/usr/local/bin`:

            sudo cp csvtk /usr/local/bin/

    1. Or copy to anywhere in the environment variable `PATH`:

            mkdir -p $HOME/bin/; cp csvtk $HOME/bin/

1. **For windows**, just copy `csvtk.exe` to `C:\WINDOWS\system32`.

#### Method 2: Install via conda (latest stable version)  [![Anaconda Cloud](	https://anaconda.org/bioconda/csvtk/badges/version.svg)](https://anaconda.org/bioconda/csvtk) [![downloads](https://anaconda.org/bioconda/csvtk/badges/downloads.svg)](https://anaconda.org/bioconda/csvtk)

    conda install -c bioconda csvtk

#### Method 3: For Go developer (latest stable/dev version)

    go get -u github.com/shenwei356/csvtk/csvtk

#### Method 4: For ArchLinux AUR users (may be not the latest)

    yaourt -S csvtk

## Bash-completion

Note: The current version supports Bash only.
This should work for *nix systems with Bash installed.

Howto:

1. run: `csvtk genautocomplete`

2. create and edit `~/.bash_completion` file if you don't have it.

        nano ~/.bash_completion

    add the following:

        for bcfile in ~/.bash_completion.d/* ; do
          . $bcfile
        done

## Compared to `csvkit`

[csvkit](http://csvkit.readthedocs.org/), attention: this table wasn't updated for 2 years.

Features                |  csvtk   |  csvkit   |   Note
:-----------------------|:--------:|:---------:|:---------
Read    Gzip            |   Yes    |  Yes      | read gzip files
Fields ranges           |   Yes    |  Yes      | e.g. `-f 1-4,6`
**Unselect fileds**     |   Yes    |  --       | e.g. `-1` for excluding first column
**Fuzzy fields**        |   Yes    |  --       | e.g. `ab*` for columns with name prefix "ab"
Reorder fields          |   Yes    |  Yes      | it means `-f 1,2` is different from `-f 2,1`
**Rename columns**      |   Yes    |  --       | rename with new name(s) or from existed names
Sort by multiple keys   |   Yes    |  Yes      | bash sort like operations
**Sort by number**      |   Yes    |  --       | e.g. `-k 1:n`
**Multiple sort**       |   Yes    |  --       | e.g. `-k 2:r -k 1:nr`
Pretty output           |   Yes    |  Yes      | convert CSV to readable aligned table
**Unique data**         |   Yes    |  --       | unique data of selected fields
**frequency**           |   Yes    |  --       | frequencies of selected fields
**Sampling**            |   Yes    |  --       | sampling by proportion
**Mutate fields**       |   Yes    |  --       | create new columns from selected fields
**Repalce**             |   Yes    |  --       | replace data of selected fields

Similar tools:

- [csvkit](http://csvkit.readthedocs.org/) - A suite of utilities for converting to and working with CSV, the king of tabular file formats. http://csvkit.rtfd.org/
- [xsv](https://github.com/BurntSushi/xsv) - A fast CSV toolkit written in Rust.
- [miller](https://github.com/johnkerl/miller) - Miller is like sed, awk, cut, join, and sort for
name-indexed data such as CSV and tabular JSON http://johnkerl.org/miller
- [tsv-utils](https://github.com/eBay/tsv-utils) - Command line utilities for tab-separated value files written in the D programming language.

## Examples

More [examples](http://shenwei356.github.io/csvtk/usage/) and [tutorial](http://shenwei356.github.io/csvtk/tutorial/).

**Attention**

1. The CSV parser requires all the lines have same number of fields/columns.
    Even lines with spaces will cause error.
    Use '-I/--ignore-illegal-row' to skip these lines if neccessary.
2. By default, csvtk thinks your files have header row, if not, switch flag `-H` on.
3. Column names better be unique.
4. By default, lines starting with `#` will be ignored, if the header row
    starts with `#`, please assign flag `-C` another rare symbol, e.g. `'$'`.
5. By default, csvtk handles CSV files, use flag `-t` for tab-delimited files.
6. If `"` exists in tab-delimited files, use flag `-l`.
7. Do not mix use digital fields and column names.

Examples

1. Pretty result

        $ csvtk pretty names.csv
        id   first_name   last_name   username
        11   Rob          Pike        rob
        2    Ken          Thompson    ken
        4    Robert       Griesemer   gri
        1    Robert       Thompson    abc
        NA   Robert       Abel        123

1. Summary of selected digital fields, supporting "group-by"

        $ cat testdata/digitals2.csv \
            | csvtk summary --ignore-non-digits --fields f4:sum,f5:sum --groups f1,f2 \
            | csvtk pretty
        f1    f2     f4:sum   f5:sum
        bar   xyz    7.00     106.00
        bar   xyz2   4.00     4.00
        foo   bar    6.00     3.00
        foo   bar2   4.50     5.00

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

1. **Rename column names** (`rename` and `rename2`)

    - Setting new names: `csvtk rename -f A,B -n a,b` or `csvtk rename -f 1-3 -n a,b,c`
    - Replacing with original names by regular express: `cat ../testdata/c.csv | ./csvtk rename2 -F -f "*" -p "(.*)" -r 'prefix_$1'` for adding prefix to all column names.

1. **Edit data with regular expression** (`replace`)

    - Remove Chinese charactors:  `csvtk replace -F -f "*_name" -p "\p{Han}+" -r ""`

1. **Create new column from selected fields by regular expression** (`mutate`)

    - In default, copy a column: `csvtk mutate -f id `
    - Extract prefix of data as group name (get "A" from "A.1" as group name):
      `csvtk mutate -f sample -n group -p "^(.+?)\."`

1. Sort by multiple keys (`sort`)

    - By single column : `csvtk sort -k 1` or `csvtk sort -k last_name`
    - By multiple columns: `csvtk sort -k 1,2` or `csvtk sort -k 1 -k 2` or `csvtk sort -k last_name,age`
    - Sort by number: `csvtk sort -k 1:n` or  `csvtk sort -k 1:nr` for reverse number
    - Complex sort: `csvtk sort -k region -k age:n -k id:nr`
    - In natural order: `csvtk sort -k chr:N`

1. **Join multiple files by keys** (`join`)

    - All files have same key column: `csvtk join -f id file1.csv file2.csv`
    - Files have different key columns: `csvtk join -f "username;username;name" names.csv phone.csv adress.csv -k`

1. Filter by numbers (`filter`)

    - Single field: `csvtk filter -f "id>0"`
    - **Multiple fields**: `csvtk filter -f "1-3>0"`
    - Using `--any` to print record if any of the field satisfy the condition: `csvtk filter -f "1-3>0" --any`
    - **fuzzy fields**: `csvtk filter -F -f "A*!=0"`

1. **Filter rows by awk-like arithmetic/string expressions** (`filter2`)

    - Using field index: `csvtk filter2 -f '$3>0'`
    - Using column names: `csvtk filter2 -f '$id > 0'`
    - Both arithmetic and string expressions: `csvtk filter2 -f '$id > 3 || $username=="ken"'`
    - More complicated: `csvtk filter2 -H -t -f '$1 > 2 && $2 % 2 == 0'`

1. Ploting
    - plot histogram with data of the second column:
     
            csvtk -t plot hist testdata/grouped_data.tsv.gz -f 2 | display

      ![histogram.png](testdata/figures/histogram.png)
        
    - plot boxplot with data of the "GC Content" (third) column,
    group information is the "Group" column.
    
            csvtk -t plot box testdata/grouped_data.tsv.gz -g "Group" \
                -f "GC Content" --width 3 | display
            
      ![boxplot.png](testdata/figures/boxplot.png)
      
    -  plot horiz boxplot with data of the "Length" (second) column,
    group information is the "Group" column.
    
            csvtk -t plot box testdata/grouped_data.tsv.gz -g "Group" -f "Length"  \
                --height 3 --width 5 --horiz --title "Horiz box plot" | display
      
      ![boxplot2.png](testdata/figures/boxplot2.png)
      
    - plot line plot with X-Y data
    
            csvtk -t plot line testdata/xy.tsv -x X -y Y -g Group | display
            
      ![lineplot.png](testdata/figures/lineplot.png)
      
    - plot scatter plot with X-Y data
        
            csvtk -t plot line testdata/xy.tsv -x X -y Y -g Group --scatter | display
            
      ![scatter.png](testdata/figures/scatter.png)

## Acknowledgements

We are grateful to [Zhiluo Deng](https://github.com/dawnmy) and
[Li Peng](https://github.com/penglbio) for suggesting features and reporting bugs.

Thanks [Albert Vilella](https://github.com/avilella) for features suggestion,
which makes csvtk feature-rich。

## Contact

[Create an issue](https://github.com/shenwei356/csvtk/issues) to report bugs,
propose new functions or ask for help.

Or [leave a comment](https://shenwei356.github.io/csvtk/usage/#disqus_thread).

## License

[MIT License](https://github.com/shenwei356/csvtk/blob/master/LICENSE)
