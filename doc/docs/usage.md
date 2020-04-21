# Usage and Examples

## Before use

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

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
## Table of Contents

- [csvtk](#csvtk)

**Information**

- [headers](#headers)
- [dim](#dim)
- [summary](#summary)
- [corr](#corr)
- [watch](#watch)

**Format conversion**

- [pretty](#pretty)
- [transpose](#transpose)
- [csv2md](#csv2md)
- [csv2json](#csv2json)
- [xlsx2csv](#xlsx2csv)

**Set operations**

- [head](#head)
- [concat](#concat)
- [sample](#sample)
- [cut](#cut)
- [grep](#grep)
- [uniq](#uniq)
- [freq](#freq)
- [inter](#inter)
- [filter](#filter)
- [filter2](#filter2)
- [join](#join)
- [split](#split)
- [splitxlsx](#splitxlsx)
- [collapse](#collapse)
- [comb](#comb)

**Edit**

- [add-header](#add-header)
- [del-header](#del-header)
- [rename](#rename)
- [rename2](#rename2)
- [replace](#replace)
- [mutate](#mutate)
- [mutate2](#mutate2)
- [sep](#sep)
- [gather](#gather)

**Ordering**

- [sort](#sort)

**Ploting**

- [plot](#plot)
- [plot hist](#plot-hist)
- [plot box](#plot-box)
- [plot line](#plot-line)

**Misc**

- [cat](#cat)
- [genautocomplete](#genautocomplete)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## csvtk

Usage

```text
csvtk -- a cross-platform, efficient and practical CSV/TSV toolkit

Version: 0.20.0

Author: Wei Shen <shenwei356@gmail.com>

Documents  : http://shenwei356.github.io/csvtk
Source code: https://github.com/shenwei356/csvtk

Attention:

  1. The CSV parser requires all the lines have same number of fields/columns.
     Even lines with spaces will cause error. 
     Use '-I/--ignore-illegal-row' to skip these lines if neccessary.
  2. By default, csvtk thinks your files have header row, if not, switch flag "-H" on.
  3. Column names better be unique.
  4. By default, lines starting with "#" will be ignored, if the header row
     starts with "#", please assign flag "-C" another rare symbol, e.g. '$'.
  5. By default, csvtk handles CSV files, use flag "-t" for tab-delimited files.
  6. If " exists in tab-delimited files, use flag "-l".
  7. Do not mix use digital fields and column names.

Environment variables for frequently used global flags

  - "CSVTK_T" for flag "-t/--tabs"
  - "CSVTK_H" for flag "-H/--no-header-row"

Usage:
  csvtk [command]

Available Commands:
  add-header      add column names
  cat             stream file to stdout and report progress on stderr
  collapse        collapse one field with selected fields as keys
  comb            compute combinations of items at every row
  concat          concatenate CSV/TSV files by rows
  corr            calculate Pearson correlation between two columns
  csv2json        convert CSV to JSON format
  csv2md          convert CSV to markdown format
  csv2tab         convert CSV to tabular format
  cut             select parts of fields
  del-header      delete column names
  dim             dimensions of CSV file
  filter          filter rows by values of selected fields with arithmetic expression
  filter2         filter rows by awk-like artithmetic/string expressions
  freq            frequencies of selected fields
  gather          gather columns into key-value pairs
  genautocomplete generate shell autocompletion script
  grep            grep data by selected fields with patterns/regular expressions
  head            print first N records
  headers         print headers
  help            Help about any command
  inter           intersection of multiple files
  join            join files by selected fields (inner, left and outer join)
  mutate          create new column from selected fields by regular expression
  mutate2         create new column from selected fields by awk-like artithmetic/string expressions
  plot            plot common figures
  pretty          convert CSV to readable aligned table
  rename          rename column names with new names
  rename2         rename column names by regular expression
  replace         replace data of selected fields by regular expression
  sample          sampling by proportion
  sep             separate column into multiple columns
  sort            sort by selected fields
  space2tab       convert space delimited format to CSV
  split           split CSV/TSV into multiple files according to column values
  splitxlsx       split XLSX sheet into multiple sheets according to column values
  summary         summary statistics of selected digital fields (groupby group fields)
  tab2csv         convert tabular format to CSV
  transpose       transpose CSV data
  uniq            unique data without sorting
  version         print version information and check for update
  watch           monitor the specified fields
  xlsx2csv        convert XLSX to CSV format

Flags:
  -c, --chunk-size int         chunk size of CSV reader (default 50)
  -C, --comment-char string    lines starting with commment-character will be ignored. if your header row starts with '#', please assign "-C" another rare symbol, e.g. '$' (default "#")
  -d, --delimiter string       delimiting character of the input CSV file (default ",")
  -h, --help                   help for csvtk
  -E, --ignore-empty-row       ignore empty rows
  -I, --ignore-illegal-row     ignore illegal rows
      --infile-list string     file of input files list (one file per line), if given, they are appended to files from cli arguments
  -l, --lazy-quotes            if given, a quote may appear in an unquoted field and a non-doubled quote may appear in a quoted field
  -H, --no-header-row          specifies that the input CSV file does not have header row
  -j, --num-cpus int           number of CPUs to use (default value depends on your computer) (default 16)
  -D, --out-delimiter string   delimiting character of the output CSV file, e.g., -D $'\t' for tab (default ",")
  -o, --out-file string        out file ("-" for stdout, suffix .gz for gzipped out) (default "-")
  -T, --out-tabs               specifies that the output is delimited with tabs. Overrides "-D"
  -t, --tabs                   specifies that the input CSV file is delimited with tabs. Overrides "-d" and "-D"
```

## headers

Usage

```text
print headers

Usage:
  csvtk headers [flags]

```

Examples

```sh
$ csvtk headers testdata/*.csv
# testdata/1.csv
1       name
2       attr
# testdata/2.csv
1       name
2       major
# testdata/3.csv
1       id
2       name
3       hobby
```

## dim

Usage

```text
dimensions of CSV file

Usage:
  csvtk dim [flags]

Aliases:
  dim, size, stats, stat

Flags:
  -h, --help   help for dim

```

Examples

1. with header row

        $ cat testdata/names.csv
        id,first_name,last_name,username
        11,"Rob","Pike",rob
        2,Ken,Thompson,ken
        4,"Robert","Griesemer","gri"
        1,"Robert","Thompson","abc"
        NA,"Robert","Abel","123"

        $ cat testdata/names.csv | csvtk size
        file   num_cols   num_rows
        -             4          5

2. no header row

        $ cat  testdata/digitals.tsv
        4       5       6
        1       2       3
        7       8       0
        8       1,000   4

        $ cat  testdata/digitals.tsv \
            | csvtk size -t -H
        file   num_cols   num_rows
        -             3          4

## summary

Usage

```text
summary statistics of selected digital fields (groupby group fields)

Attention:

  1. Do not mix use digital fields and column names.

Available operations:
 
  # numeric/statistical operations
  # provided by github.com/gonum/stat and github.com/gonum/floats
  countn (count of digits), min, max, sum,
  mean, stdev, variance, median, q1, q2, q3,
  entropy (Shannon entropy),
  prod (product of the elements)

  # textual/numeric operations
  count, first, last, rand, unique, collapse, countunique

Usage:
  csvtk summary [flags]

Flags:
  -n, --decimal-width int   limit floats to N decimal points (default 2)
  -f, --fields strings      operations on these fields. e.g -f 1:count,1:sum or -f colA:mean. available operations: collapse, count, countn, countunique, entropy, first, last, max, mean, median, min, prod, q1, q2, q3, rand, stdev, sum, uniq, variance
  -g, --groups string       group via fields. e.g -f 1,2 or -f columnA,columnB
  -h, --help                help for summary
  -i, --ignore-non-digits   ignore non-digital values like "NA" or "N/A"
  -S, --rand-seed int       rand seed for operation "rand" (default 11)
  -s, --separater string    separater for collapsed data (default "; ")

```

Examples

1. data

        $ cat testdata/digitals2.csv 
        f1,f2,f3,f4,f5
        foo,bar,xyz,1,0
        foo,bar2,xyz,1.5,-1
        foo,bar2,xyz,3,2
        foo,bar,xyz,5,3
        foo,bar2,xyz,N/A,4
        bar,xyz,abc,NA,2
        bar,xyz,abc2,1,-1
        bar,xyz,abc,2,0
        bar,xyz,abc,1,5
        bar,xyz,abc,3,100
        bar,xyz2,abc3,2,3
        bar,xyz2,abc3,2,1

1. use flag `-i/--ignore-non-digits`

        $ cat testdata/digitals2.csv \
            | csvtk summary -f f4:sum
        [ERRO] column 4 has non-digital data: N/A, you can use flag -i/--ignore-non-digits to skip these data

        $ cat testdata/digitals2.csv \
            | csvtk summary -f f4:sum -i
        f4:sum
        21.50

1. multiple fields suported

        $ cat testdata/digitals2.csv \
            | csvtk summary -f f4:sum,f5:sum -i
        f4:sum,f5:sum
        21.50,118.00

1. using fields instead of colname is still supported

        $ cat testdata/digitals2.csv \
            | csvtk summary -f 4:sum,5:sum -i
        f4:sum,f5:sum
        21.50,118.00

1. but remember not mixing use digital fields and column names

        $ cat testdata/digitals2.csv \
            | csvtk summary -f f4:sum,5:sum -i
        [ERRO] column "5" not existed in file: -

        $ cat testdata/digitals2.csv \
            | csvtk summary -f 4:sum,f5:sum -i
        [ERRO] fail to parse digital field: f5, you may mix use digital fields and column names

1. groupby

        $ cat testdata/digitals2.csv \
            | csvtk summary -i -f f4:sum,f5:sum -g f1,f2 \
            | csvtk pretty
        f1    f2     f4:sum   f5:sum
        bar   xyz    7.00     106.00
        bar   xyz2   4.00     4.00
        foo   bar    6.00     3.00
        foo   bar2   4.50     5.00

1. for data without header line

        $ cat testdata/digitals2.csv | sed 1d \
            | csvtk summary -H -i -f 4:sum,5:sum -g 1,2 \
            | csvtk pretty
        bar   xyz    7.00   106.00
        bar   xyz2   4.00   4.00
        foo   bar    6.00   3.00
        foo   bar2   4.50   5.00

1. numeric/statistical operations

        $ cat testdata/digitals2.csv \
            | csvtk summary -i -g f1 -f f4:countn,f4:mean,f4:stdev,f4:q1,f4:q2,f4:mean,f4:q3,f4:min,f4:max \
            | csvtk pretty
        f1    f4:countn   f4:mean   f4:stdev   f4:q1   f4:q2   f4:mean   f4:q3   f4:min   f4:max
        bar   6.00        1.83      0.75       1.00    2.00    1.83      2.00    1.00     3.00
        foo   4.00        2.62      1.80       1.25    2.25    2.62      4.00    1.00     5.00

1. textual/numeric operations

        $  cat testdata/digitals2.csv \
            | csvtk summary -i -g f1 -f f2:count,f2:first,f2:last,f2:rand,f2:collapse,f2:uniq,f2:countunique \
            | csvtk pretty
        f1    f2:count   f2:first   f2:last   f2:rand   f2:collapse                           f2:uniq     f2:countunique
        bar   7          xyz        xyz2      xyz2      xyz; xyz; xyz; xyz; xyz; xyz2; xyz2   xyz2; xyz   2
        foo   5          bar        bar2      bar2      bar; bar2; bar2; bar; bar2            bar; bar2   2

1. mixed operations

        $  cat testdata/digitals2.csv \
            | csvtk summary -i -g f1 -f f4:collapse,f4:max \
            | csvtk pretty
        f1    f4:collapse            f4:max
        bar   NA; 1; 2; 1; 3; 2; 2   3.00
        foo   1; 1.5; 3; 5; N/A      5.00

1. `count` and `countn` (count of digits)

        $ cat testdata/digitals2.csv \
            | csvtk summary -f f4:count,f4:countn -i \
            | csvtk pretty
        f4:count   f4:countn
        12         10
        
        # details:
        $ cat testdata/digitals2.csv \
            | csvtk summary -f f4:count,f4:countn,f4:collapse -i -g f1 \
            | csvtk pretty
        f1    f4:count   f4:countn   f4:collapse
        bar   7          6           NA; 1; 2; 1; 3; 2; 2
        foo   5          4           1; 1.5; 3; 5; N/A


## watch

Usage

```text
monitor the specified fields

Usage:
  csvtk watch [flags]

Flags:
  -B, --bins int         number of histogram bins (default -1)
  -W, --delay int        sleep this many seconds after plotting (default 1)
  -y, --dump             print histogram data to stderr instead of plotting
  -f, --field string     field to watch
  -h, --help             help for watch
  -O, --image string     save histogram to this PDF/image file
  -L, --log              log10(x+1) transform numeric values
  -x, --pass             passthrough mode (forward input to output)
  -p, --print-freq int   print/report after this many records (-1 for print after EOF) (default -1)
  -Q, --quiet            supress all plotting to stderr
  -R, --reset            reset histogram after every report
```

Examples

1. Read whole file, plot histogram of field on the terminal and PDF

        csvtk -t watch -O hist.pdf -f MyField input.tsv

1. Monitor a TSV stream, print histogram every 1000 records

        cat input.tsv | csvtk -t watch -f MyField -p 1000 -

1. Monitor a TSV stream, print histogram every 1000 records, hang forever for updates

        tail -f +0 input.tsv | csvtk -t watch -f MyField -p 1000 -

## corr

Usage

```text
calculate Pearson correlation between two columns

Usage:
  csvtk corr [flags]

Flags:
  -f, --fields string   comma separated fields
  -h, --help            help for corr
  -i, --ignore_nan      Ignore non-numeric fields to avoid returning NaN
  -L, --log             Calcute correlations on Log10 transformed data
  -x, --pass            passthrough mode (forward input to output)
```

Examples

1. Calculate pairwise correlations between field, ignore non-numeric values

        csvtk -t corr -i -f 1,Foo,Bar input.tsv


## pretty

Usage

```text
convert CSV to readable aligned table

Attention:

  pretty treats the first row as header line and requires them to be unique

Usage:
  csvtk pretty [flags]

Flags:
  -r, --align-right        align right
  -h, --help               help for pretty
  -W, --max-width int      max width
  -w, --min-width int      min width
  -s, --separator string   fields/columns separator (default "   ")

```

Examples:

1. default

        $ csvtk pretty testdata/names.csv
        id   first_name   last_name   username
        11   Rob          Pike        rob
        2    Ken          Thompson    ken
        4    Robert       Griesemer   gri
        1    Robert       Thompson    abc
        NA   Robert       Abel        123

2. align right

        $ csvtk pretty testdata/names.csv -r
        id   first_name   last_name   username
        11          Rob        Pike        rob
         2          Ken    Thompson        ken
         4       Robert   Griesemer        gri
         1       Robert    Thompson        abc
        NA       Robert        Abel        123


3. custom separator

        $ csvtk pretty testdata/names.csv -s " | "
        id | first_name | last_name | username
        11 | Rob        | Pike      | rob
        2  | Ken        | Thompson  | ken
        4  | Robert     | Griesemer | gri
        1  | Robert     | Thompson  | abc
        NA | Robert     | Abel      | 123

## transpose

Usage

```text
transpose CSV data

Usage:
  csvtk transpose [flags]

```

Examples

    $ cat  testdata/digitals.tsv
    4       5       6
    1       2       3
    7       8       0
    8       1,000   4

    $ csvtk transpose -t  testdata/digitals.tsv
    4       1       7       8
    5       2       8       1,000
    6       3       0       4

## csv2json

Usage

```text
convert CSV to JSON format

Usage:
  csvtk csv2json [flags]

Flags:
  -h, --help            help for csv2json
  -i, --indent string   indent. if given blank, output json in one line. (default "  ")
  -k, --key string      output json as an array of objects keyed by a given filed rather than as a list. e.g -k 1 or -k columnA

```

Examples

- test data

        $ cat data.csv
        ID,room,name
        3,G13,Simon
        5,103,Anna

- default operation

        $ cat data.csv | csvtk csv2json
        [
          {
            "ID": "3",
            "room": "G13",
            "name": "Simon"
          },
          {
            "ID": "5",
            "room": "103",
            "name": "Anna"
          }
        ]

- change indent

        $ cat data.csv | csvtk csv2json -i "    "
        [
            {
                "ID": "3",
                "room": "G13",
                "name": "Simon"
            },
            {
                "ID": "5",
                "room": "103",
                "name": "Anna"
            }
        ]

- change indent 2)

        $ cat data.csv | csvtk csv2json -i ""
        [{"ID":"3","room":"G13","name":"Simon"},{"ID":"5","room":"103","name":"Anna"}]

- output json as an array of objects keyed by a given filed rather than as a list.

        $ cat data.csv | csvtk csv2json -k ID
        {
          "3": {
            "ID": "3",
            "room": "G13",
            "name": "Simon"
          },
          "5": {
            "ID": "5",
            "room": "103",
            "name": "Anna"
          }
        }

- for CSV without header row

        $ cat data.csv | csvtk csv2json -H
        [
          [
            "ID",
            "room",
            "name"
          ],
          [
            "3",
            "G13",
            "Simon"
          ],
          [
            "5",
            "103",
            "Anna"
          ]
        ]

- for CSV without header row 2)

        $ cat data.csv | csvtk csv2json -H -k 1
        {
          "ID": [
            "ID",
            "room",
            "name"
          ],
          "3": [
            "3",
            "G13",
            "Simon"
          ],
          "5": [
            "5",
            "103",
            "Anna"
          ]
        }

## csv2md

Usage

```text
convert CSV to markdown format

Attention:

  csv2md treats the first row as header line and requires them to be unique

Usage:
  csvtk csv2md [flags]

Flags:
  -a, --alignments string   comma separated alignments. e.g. -a l,c,c,c or -a c (default "l")
  -w, --min-width int       min width (at least 3) (default 3)

```

Examples

1. give single alignment symbol

        $ cat testdata/names.csv | csvtk csv2md -a left
        id |first_name|last_name|username
        :--|:---------|:--------|:-------
        11 |Rob       |Pike     |rob
        2  |Ken       |Thompson |ken
        4  |Robert    |Griesemer|gri
        1  |Robert    |Thompson |abc
        NA |Robert    |Abel     |12

    result:

    id |first_name|last_name|username
    :--|:---------|:--------|:-------
    11 |Rob       |Pike     |rob
    2  |Ken       |Thompson |ken
    4  |Robert    |Griesemer|gri
    1  |Robert    |Thompson |abc
    NA |Robert    |Abel     |12

2. give alignment symbols of all fields

        $ cat testdata/names.csv | csvtk csv2md -a c,l,l,l
        id |first_name|last_name|username
        :-:|:---------|:--------|:-------
        11 |Rob       |Pike     |rob
        2  |Ken       |Thompson |ken
        4  |Robert    |Griesemer|gri
        1  |Robert    |Thompson |abc
        NA |Robert    |Abel     |123

    result

    id |first_name|last_name|username
    :-:|:---------|:--------|:-------
    11 |Rob       |Pike     |rob
    2  |Ken       |Thompson |ken
    4  |Robert    |Griesemer|gri
    1  |Robert    |Thompson |abc
    NA |Robert    |Abel     |123

## xlsx2csv

Usage

```text
convert XLSX to CSV format

Usage:
  csvtk xlsx2csv [flags]

Flags:
  -h, --help                help for xlsx2csv
  -a, --list-sheets         list all sheets
  -i, --sheet-index int     Nth sheet to retrieve (default 1)
  -n, --sheet-name string   sheet to retrieve

```

Examples

1. list all sheets

        $ csvtk xlsx2csv ../testdata/accounts.xlsx -a
        index   sheet
        1       names
        2       phones
        3       region

1. retrieve sheet by index

        $ csvtk xlsx2csv ../testdata/accounts.xlsx -i 3
        name,region
        ken,nowhere
        gri,somewhere
        shenwei,another
        Thompson,there

1. retrieve sheet by name

        $ csvtk xlsx2sv ../testdata/accounts.xlsx -n region
        name,region
        ken,nowhere
        gri,somewhere
        shenwei,another
        Thompson,there

## head

Usage

```text
print first N records

Usage:
  csvtk head [flags]

Flags:
  -n, --number int   print first N records (default 10)

```

Examples

1. with header line

        $ csvtk head -n 2 testdata/1.csv
        name,attr
        foo,cool
        bar,handsome

2. no header line

        $ csvtk head -H -n 2 testdata/1.csv
        name,attr
        foo,cool

## concat

Usage

```text
concatenate CSV/TSV files by rows

Note that the second and later files are concatenated to the first one,
so only columns match that of the first files kept.

Usage:
  csvtk concat [flags]

Flags:
  -h, --help                    help for concat
  -i, --ignore-case             ignore case (column name)
  -k, --keep-unmatched          keep blanks even if no any data of a file matches
  -u, --unmatched-repl string   replacement for unmatched data

```

Examples

1. data

        $ csvtk pretty names.csv
        id   first_name   last_name   username
        11   Rob          Pike        rob
        2    Ken          Thompson    ken
        4    Robert       Griesemer   gri
        1    Robert       Thompson    abc
        NA   Robert       Abel        123

        $ csvtk pretty  names.reorder.csv
        last_name   username   id   first_name
        Pike        rob        11   Rob
        Thompson    ken        2    Ken
        Griesemer   gri        4    Robert
        Thompson    abc        1    Robert
        Abel        123        NA   Robert

        $ csvtk pretty  names.with-unmatched-colname.csv
        id2   First_name   Last_name    Username   col
        22    Rob33        Pike222      rob111     abc
        44    Ken33        Thompson22   ken111     def

1. simple one

        $ csvtk concat names.csv names.reorder.csv \
            | csvtk pretty
        id   first_name   last_name   username
        11   Rob          Pike        rob
        2    Ken          Thompson    ken
        4    Robert       Griesemer   gri
        1    Robert       Thompson    abc
        NA   Robert       Abel        123
        11   Rob          Pike        rob
        2    Ken          Thompson    ken
        4    Robert       Griesemer   gri
        1    Robert       Thompson    abc
        NA   Robert       Abel        123

1. data with unmatched column names, and ignoring cases

        $ csvtk concat names.csv names.with-unmatched-colname.csv -i \
            | csvtk pretty
        id   first_name   last_name    username
        11   Rob          Pike         rob
        2    Ken          Thompson     ken
        4    Robert       Griesemer    gri
        1    Robert       Thompson     abc
        NA   Robert       Abel         123
             Rob33        Pike222      rob111
             Ken33        Thompson22   ken111

         $ csvtk concat names.csv names.with-unmatched-colname.csv -i -u Unmached \
            | csvtk pretty
         id         first_name   last_name    username
         11         Rob          Pike         rob
         2          Ken          Thompson     ken
         4          Robert       Griesemer    gri
         1          Robert       Thompson     abc
         NA         Robert       Abel         123
         Unmached   Rob33        Pike222      rob111
         Unmached   Ken33        Thompson22   ken111

1. Sometimes data of one file does not matche any column, they are discared by default.
  But you can keep them using flag `-k/--keep-unmatched`

        $ csvtk concat names.with-unmatched-colname.csv names.csv \
            | csvtk pretty
        id2   First_name   Last_name    Username   col
        22    Rob33        Pike222      rob111     abc
        44    Ken33        Thompson22   ken111     def

        $ csvtk concat names.with-unmatched-colname.csv names.csv -u -k NA \
            | csvtk pretty
        id2   First_name   Last_name    Username   col
        22    Rob33        Pike222      rob111     abc
        44    Ken33        Thompson22   ken111     def
        NA    NA           NA           NA         NA
        NA    NA           NA           NA         NA
        NA    NA           NA           NA         NA
        NA    NA           NA           NA         NA
        NA    NA           NA           NA         NA

## sample

Usage

```text
sampling by proportion

Usage:
  csvtk sample [flags]

Flags:
  -h, --help               help for sample
  -n, --line-number        print line number as the first column ("n")
  -p, --proportion float   sample by proportion
  -s, --rand-seed int      rand seed (default 11)

```

Examples

```sh
$ seq 100 | csvtk sample -H -p 0.5 | wc -l
46

$ seq 100 | csvtk sample -H -p 0.5 | wc -l
46

$ seq 100 | csvtk sample -H -p 0.1 | wc -l
10

$ seq 100 | csvtk sample -H -p 0.05 -n
50,50
52,52
65,65
```

## cut

Usage

```text
select parts of fields

Usage:
  csvtk cut [flags]

Flags:
  -f, --fields string   select only these fields. e.g -f 1,2 or -f columnA,columnB, or -f -columnA for unselect columnA
  -F, --fuzzy-fields    using fuzzy fields, e.g., -F -f "*name" or -F -f "id123*"
  -h, --help            help for cut
  -i, --ignore-case     ignore case (column name)

```

Examples

- data:

        $ cat testdata/names.csv
        id,first_name,last_name,username
        11,"Rob","Pike",rob
        2,Ken,Thompson,ken
        4,"Robert","Griesemer","gri"
        1,"Robert","Thompson","abc"
        NA,"Robert","Abel","123"

- Select columns by column index: `csvtk cut -f 1,2`

        $ cat testdata/names.csv \
            | csvtk cut -f 1,2
        id,first_name
        11,Rob
        2,Ken
        4,Robert
        1,Robert
        NA,Robert

- Select columns by column names: `csvtk cut -f first_name,username`

        $ cat testdata/names.csv \
            | csvtk cut -f first_name,username
        first_name,username
        Rob,rob
        Ken,ken
        Robert,gri
        Robert,abc
        Robert,123

- **Unselect**:
    - select 3+ columns: `csvtk cut -f -1,-2`

            $ cat testdata/names.csv \
                | csvtk cut -f -1,-2
            last_name,username
            Pike,rob
            Thompson,ken
            Griesemer,gri
            Thompson,abc
            Abel,123

    - select columns except `first_name`: `csvtk cut -f -first_name`

            $ cat testdata/names.csv \
                | csvtk cut -f -first_name
            id,last_name,username
            11,Pike,rob
            2,Thompson,ken
            4,Griesemer,gri
            1,Thompson,abc
            NA,Abel,123

- **Fuzzy fields** using wildcard character,  `csvtk cut -F -f "*_name,username"`

        $ cat testdata/names.csv \
            | csvtk cut -F -f "*_name,username"
        first_name,last_name,username
        Rob,Pike,rob
        Ken,Thompson,ken
        Robert,Griesemer,gri
        Robert,Thompson,abc
        Robert,Abel,123

- All fields: `csvtk cut -F -f "*"` (only works when all colnames are unique)

        $ cat testdata/names.csv \
            | csvtk cut -F -f "*"
        id,first_name,last_name,username
        11,Rob,Pike,rob
        2,Ken,Thompson,ken
        4,Robert,Griesemer,gri
        1,Robert,Thompson,abc
        NA,Robert,Abel,123

- Field ranges:
    - `csvtk cut -f 2-4` for column 2,3,4

            $ cat testdata/names.csv \
                | csvtk cut -f 2-4
            first_name,last_name,username
            Rob,Pike,rob
            Ken,Thompson,ken
            Robert,Griesemer,gri
            Robert,Thompson,abc
            Robert,Abel,123

    - `csvtk cut -f -3--1` for discarding column 1,2,3

            $ cat testdata/names.csv \
                | csvtk cut -f -3--1
            username
            rob
            ken
            gri
            abc
            123

## uniq

Usage

```text
unique data without sorting

Usage:
  csvtk uniq [flags]

Flags:
  -f, --fields string   select only these fields. e.g -f 1,2 or -f columnA,columnB (default "1")
  -F, --fuzzy-fields    using fuzzy fields, e.g., -F -f "*name" or -F -f "id123*"
  -i, --ignore-case     ignore case

```

Examples:

- data:

        $ cat testdata/names.csv
        id,first_name,last_name,username
        11,"Rob","Pike",rob
        2,Ken,Thompson,ken
        4,"Robert","Griesemer","gri"
        1,"Robert","Thompson","abc"
        NA,"Robert","Abel","123"

- unique first_name (it removes rows with duplicated first_name)

        $ cat testdata/names.csv \
            | csvtk uniq -f first_name
        id,first_name,last_name,username
        11,Rob,Pike,rob
        2,Ken,Thompson,ken
        4,Robert,Griesemer,gri

- unique first_name, a more common way

        $ cat testdata/names.csv \
            | csvtk cut -f first_name \
            | csvtk uniq -f 1
        first_name
        Rob
        Ken
        Robert

## freq

Usage

```text
frequencies of selected fields

Usage:
  csvtk freq [flags]

Flags:
  -f, --fields string   select only these fields. e.g -f 1,2 or -f columnA,columnB (default "1")
  -F, --fuzzy-fields    using fuzzy fields, e.g., -F -f "*name" or -F -f "id123*"
  -i, --ignore-case     ignore case
  -r, --reverse         reverse order while sorting
  -n, --sort-by-freq    sort by frequency
  -k, --sort-by-key     sort by key

```

Examples

1. one filed

        $ cat testdata/names.csv \
            | csvtk freq -f first_name | csvtk pretty
        first_name   frequency
        Ken          1
        Rob          1
        Robert       3

1. sort by frequency. you can also use `csvtk sort` with more sorting options

        $ cat testdata/names.csv \
            | csvtk freq -f first_name -n -r \
            | csvtk pretty
        first_name   frequency
        Robert       3
        Ken          1
        Rob          1

1. sorty by key

        $ cat testdata/names.csv \
            | csvtk freq -f first_name -k \
            | csvtk pretty
        first_name   frequency
        Ken          1
        Rob          1
        Robert       3

1. multiple fields

        $ cat testdata/names.csv \
            | csvtk freq -f first_name,last_name \
            | csvtk pretty
        first_name   last_name   frequency
        Robert       Abel        1
        Ken          Thompson    1
        Rob          Pike        1
        Robert       Thompson    1
        Robert       Griesemer   1

1. data without header row

        $ cat testdata/ testdata/digitals.tsv \
            | csvtk -t -H freq -f 1
        8       1
        1       1
        4       1
        7       1

## inter

Usage

```text
intersection of multiple files

Attention:

  1. fields in all files should be the same, 
     if not, extracting to another file using "csvtk cut".

Usage:
  csvtk inter [flags]

Flags:
  -f, --fields string   select only these fields. e.g -f 1,2 or -f columnA,columnB (default "1")
  -F, --fuzzy-fields    using fuzzy fields, e.g., -F -f "*name" or -F -f "id123*"
  -i, --ignore-case     ignore case

```

Examples:

    $ cat testdata/phones.csv
    username,phone
    gri,11111
    rob,12345
    ken,22222
    shenwei,999999

    $ cat testdata/region.csv
    name,region
    ken,nowhere
    gri,somewhere
    shenwei,another
    Thompson,there

    $ csvtk inter testdata/phones.csv testdata/region.csv
    username
    gri
    ken
    shenwei

## grep

Usage

```text
grep data by selected fields with patterns/regular expressions

Usage:
  csvtk grep [flags]

Flags:
      --delete-matched        delete a pattern right after being matched, this keeps the firstly matched data and speedups when using regular expressions
  -f, --fields string         comma separated key fields, column name or index. e.g. -f 1-3 or -f id,id2 or -F -f "group*" (default "1")
  -F, --fuzzy-fields          using fuzzy fields, e.g., -F -f "*name" or -F -f "id123*"
  -h, --help                  help for grep
  -i, --ignore-case           ignore case
  -v, --invert                invert match
  -n, --line-number           print line number as the first column ("n")
  -N, --no-highlight          no highlight
  -p, --pattern strings       query pattern (multiple values supported)
  -P, --pattern-file string   pattern files (one pattern per line)
  -r, --use-regexp            patterns are regular expression
      --verbose               verbose output

```

Examples

Matched parts will be ***highlight***.

- By exact keys

        $ cat testdata/names.csv \
            | csvtk grep -f last_name -p Pike -p Abel \
            | csvtk pretty
        id   first_name   last_name   username
        11   Rob          Pike        rob
        NA   Robert       Abel        123
        
        # another form of multiple keys 
        $ csvtk grep -f last_name -p Pike,Abel,Tom

- By regular expression: `csvtk grep -f first_name -r -p Rob`

        $ cat testdata/names.csv \
            | csvtk grep -f first_name -r -p Rob \
            | csvtk pretty
        id   first_name   last_name   username
        11   Rob          Pike        rob
        4    Robert       Griesemer   gri
        1    Robert       Thompson    abc
        NA   Robert       Abel        123

- By pattern list

        $ csvtk grep -f first_name -P name_list.txt
        
- Remore rows containing any missing data (NA): 

        $ csvtk grep -F -f "*" -r -p "^$" -v 
        
- Show line number

        $ cat names.csv \
            | csvtk pretty
        id   first_name   last_name   username
        11   Rob          Pike        rob
        2    Ken          Thompson    ken
        4    Robert       Griesemer   gri
        1    Robert       Thompson    abc
        NA   Robert       Abel        123

        $ cat names.csv \
            | csvtk grep -f first_name -r -i -p rob -n \
            | csvtk pretty
        n   id   first_name   last_name   username
        1   11   Rob          Pike        rob
        3   4    Robert       Griesemer   gri
        4   1    Robert       Thompson    abc
        5   NA   Robert       Abel        123

## filter

Usage

```text
filter rows by values of selected fields with arithmetic expression

Usage:
  csvtk filter [flags]

Flags:
      --any             print record if any of the field satisfy the condition
  -f, --filter string   filter condition. e.g. -f "age>12" or -f "1,3<=2" or -F -f "c*!=0"
  -F, --fuzzy-fields    using fuzzy fields, e.g., -F -f "*name" or -F -f "id123*"
  -h, --help            help for filter
  -n, --line-number     print line number as the first column ("n")

```

Examples

1. single field

        $ cat testdata/names.csv
        id,first_name,last_name,username
        11,"Rob","Pike",rob
        2,Ken,Thompson,ken
        4,"Robert","Griesemer","gri"
        1,"Robert","Thompson","abc"
        NA,"Robert","Abel","123"

        $ cat testdata/names.csv \
            | csvtk filter -f "id>0" \
            | csvtk pretty
        id   first_name   last_name   username
        11   Rob          Pike        rob
        2    Ken          Thompson    ken
        4    Robert       Griesemer   gri
        1    Robert       Thompson    abc

2. multiple fields

        $ cat  testdata/digitals.tsv
        4       5       6
        1       2       3
        7       8       0
        8       1,000   4

        $ cat  testdata/digitals.tsv \
            | csvtk -t -H filter -f "1-3>0"
        4       5       6
        1       2       3
        8       1,000   4

    using `--any` to print record if any of the field satisfy the condition

        $  cat  testdata/digitals.tsv \
            | csvtk -t -H filter -f "1-3>0" --any
        4       5       6
        1       2       3
        7       8       0
        8       1,000   4

3. fuzzy fields

        $  cat testdata/names.csv \
            | csvtk filter -F -f "i*!=0"
        id,first_name,last_name,username
        11,Rob,Pike,rob
        2,Ken,Thompson,ken
        4,Robert,Griesemer,gri
        1,Robert,Thompson,abc

## filter2

Usage

```text
filter rows by awk-like artithmetic/string expressions

The artithmetic/string expression is supported by:

  https://github.com/Knetic/govaluate

Supported operators and types:

  Modifiers: + - / * & | ^ ** % >> <<
  Comparators: > >= < <= == != =~ !~
  Logical ops: || &&
  Numeric constants, as 64-bit floating point (12345.678)
  String constants (single quotes: 'foobar')
  Date constants (single quotes)
  Boolean constants: true false
  Parenthesis to control order of evaluation ( )
  Arrays (anything separated by , within parenthesis: (1, 2, 'foo'))
  Prefixes: ! - ~
  Ternary conditional: ? :
  Null coalescence: ??

Usage:
  csvtk filter2 [flags]

Flags:
  -f, --filter string   awk-like filter condition. e.g. '$age>12' or '$1 > $3' or '$name=="abc"' or '$1 % 2 == 0'
  -h, --help            help for filter2
  -n, --line-number     print line number as the first column ("n")

```

Examples:

1. filter rows with `id` greater than 3:

        $ cat testdata/names.csv
        id,first_name,last_name,username
        11,"Rob","Pike",rob
        2,Ken,Thompson,ken
        4,"Robert","Griesemer","gri"
        1,"Robert","Thompson","abc"
        NA,"Robert","Abel","123"

        $ cat testdata/names.csv \
            | csvtk filter2 -f '$id > 3'
        id,first_name,last_name,username
        11,Rob,Pike,rob
        4,Robert,Griesemer,gri

1. arithmetic and string expressions

        $ cat testdata/names.csv \
            | csvtk filter2 -f '$id > 3 || $username=="ken"'
        id,first_name,last_name,username
        11,Rob,Pike,rob
        2,Ken,Thompson,ken
        4,Robert,Griesemer,gri

1. More arithmetic expressions

        $ cat testdata/digitals.tsv
        4       5       6
        1       2       3
        7       8       0
        8       1,000   4

        $ cat testdata/digitals.tsv \
            | csvtk filter2 -H -t -f '$1 > 2 && $2 % 2 == 0'
        7       8       0
        8       1,000   4

        # comparison between fields and support
        $ cat testdata/digitals.tsv \
            | csvtk filter2 -H -t -f '$2 <= $3 || ( $1 / $2 > 0.5 )'
        4       5       6
        1       2       3
        7       8       0

## join

Usage

```text
join files by selected fields (inner, left and outer join).

Attention:

  1. Multiple keys supported, but the orders are ignored.
  2. Default operation is inner join, use --left-join for left join 
     and --outer-join for outer join.

Usage:
  csvtk join [flags]

Aliases:
  join, merge

Flags:
  -f, --fields string    Semicolon separated key fields of all files, if given one, we think all the files have the same key columns. Fields of different files should be separated by ";", e.g -f "1;2" or -f "A,B;C,D" or -f id (default "1")
  -F, --fuzzy-fields     using fuzzy fields, e.g., -F -f "*name" or -F -f "id123*"
  -h, --help             help for join
  -i, --ignore-case      ignore case
  -k, --keep-unmatched   keep unmatched data of the first file (left join)
  -L, --left-join        left join, equals to -k/--keep-unmatched, exclusive with --outer-join
      --na string        content for filling NA data
  -O, --outer-join       outer join, exclusive with --left-join

```

Examples:

- data

        $ cat testdata/phones.csv
        username,phone
        gri,11111
        rob,12345
        ken,22222
        shenwei,999999

        $ cat testdata/region.csv
        name,region
        ken,nowhere
        gri,somewhere
        shenwei,another
        Thompson,there


- All files have same key column: `csvtk join -f id file1.csv file2.csv`

        $ csvtk join -f 1 testdata/phones.csv testdata/region.csv \
            | csvtk pretty
        username   phone    region
        gri        11111    somewhere
        ken        22222    nowhere
        shenwei    999999   another

- keep unmatched (left join)

        $ csvtk join -f 1 testdata/phones.csv testdata/region.csv --left-join \
            | csvtk pretty
        username   phone    region
        gri        11111    somewhere
        rob        12345    
        ken        22222    nowhere
        shenwei    999999   another


- keep unmatched and fill with something

        $ csvtk join -f 1 testdata/phones.csv testdata/region.csv --left-join --na NA \
            | csvtk pretty
        username   phone    region
        gri        11111    somewhere
        rob        12345    NA
        ken        22222    nowhere
        shenwei    999999   another

- Outer join

        $ csvtk join -f 1 testdata/phones.csv testdata/region.csv --outer-join --na NA \
            | csvtk pretty
        username   phone    region
        gri        11111    somewhere
        rob        12345    NA
        ken        22222    nowhere
        shenwei    999999   another
        Thompson   NA       there

- Files have different key columns: `csvtk join -f "username;username;name" testdata/names.csv phone.csv adress.csv -k`. ***Note that fields are separated with `;` not `,`.***

        $ csvtk join -f "username;name"  testdata/phones.csv testdata/region.csv --left-join --na NA \
            | csvtk pretty
        username   phone    region
        gri        11111    somewhere
        rob        12345    NA
        ken        22222    nowhere
        shenwei    999999   another
        


- Some special cases

        $ cat testdata/1.csv
        name,attr
        foo,cool
        bar,handsome
        bob,beutiful

        $ cat testdata/2.csv
        name,major
        bar,bioinformatics
        bob,microbiology
        bob,computer science

        $ cat testdata/3.csv
        id,name,hobby
        1,bar,baseball
        2,bob,basketball
        3,foo,football
        4,wei,programming

        # nothing special
        $ csvtk join testdata/{1,2,3}.csv -f name --outer-join --na NA \
            | csvtk pretty
        name   attr       major               id   hobby
        foo    cool       NA                  3    football
        bar    handsome   bioinformatics      1    baseball
        bob    beutiful   microbiology        2    basketball
        bob    beutiful   computer science    2    basketball
        wei    NA         NA                  4    programming
        
        # just reorder files
        $ csvtk join testdata/{3,2,1}.csv -f name --outer-join --na NA \
            | csvtk pretty
        id   name   hobby         major               attr
        1    bar    baseball      bioinformatics      handsome
        2    bob    basketball    microbiology        beutiful
        2    bob    basketball    computer science    beutiful
        3    foo    football      NA                  cool
        4    wei    programming   NA                  NA
        
        # special case: names in 3.csv contain all names in all files
        $ csvtk join testdata/{3,2,1}.csv -f name --left-join --na NA \
            | csvtk pretty
        id   name   hobby         major               attr
        1    bar    baseball      bioinformatics      handsome
        2    bob    basketball    microbiology        beutiful
        2    bob    basketball    computer science    beutiful
        3    foo    football      NA                  cool
        4    wei    programming   NA                  NA


## split

Usage

```text
split CSV/TSV into multiple files according to column values

Note:

  1. flag -o/--out-file can specify out directory for splitted files

Usage:
  csvtk split [flags]

Flags:
  -g, --buf-groups int   buffering N groups before writing to file (default 100)
  -b, --buf-rows int     buffering N rows for every group before writing to file (default 100000)
  -f, --fields string    comma separated key fields, column name or index. e.g. -f 1-3 or -f id,id2 or -F -f "group*" (default "1")
  -F, --fuzzy-fields     using fuzzy fields, e.g., -F -f "*name" or -F -f "id123*"
  -h, --help             help for split
  -i, --ignore-case      ignore case
  -G, --out-gzip         force output gzipped file

```

Examples

1. Test data

        $ cat names.csv
        id,first_name,last_name,username
        11,"Rob","Pike",rob
        2,Ken,Thompson,ken
        4,"Robert","Griesemer","gri"
        1,"Robert","Thompson","abc"
        NA,"Robert","Abel","123"

1. split according to `first_name`

        $ csvtk split names.csv -f first_name
        $ ls *.csv
        names.csv  names-Ken.csv  names-Rob.csv  names-Robert.csv

        $ cat names-Ken.csv
        id,first_name,last_name,username
        2,Ken,Thompson,ken

        $ cat names-Rob.csv
        id,first_name,last_name,username
        11,Rob,Pike,rob

        $ cat names-Robert.csv
        id,first_name,last_name,username
        4,Robert,Griesemer,gri
        1,Robert,Thompson,abc
        NA,Robert,Abel,123

1. split according to `first_name` and `last_name`

        $ csvtk split names.csv -f first_name,last_name
        $ ls *.csv
        names.csv               names-Robert-Abel.csv       names-Robert-Thompson.csv
        names-Ken-Thompson.csv  names-Robert-Griesemer.csv  names-Rob-Pike.csv

1.  flag `-o/--out-file` can specify out directory for splitted files

        $ seq 10000 | csvtk split -H -o result
        $ ls result/*.csv | wc -l
        10000

1. extreme example 1: lots (1M) of rows in groups

        $ yes 2 | head -n 10000000 | gzip -c > t.gz

        $ memusg -t csvtk -H split t.gz
        elapsed time: 7.959s
        peak rss: 35.7 MB

        # check
        $ zcat t-2.gz | wc -l
        10000000
        $ zcat t-2.gz | md5sum
        f194afd7cecf645c0e3cce50c9bc526e  -
        $ zcat t.gz | md5sum
        f194afd7cecf645c0e3cce50c9bc526e  -

1. extreme example 2: lots (10K) of groups

        $ seq 10000 | gzip -c > t2.gz

        $ memusg -t csvtk -H split t2.gz  -o t2
        elapsed time: 20.856s
        peak rss: 23.77 MB

        # check
        $ ls t2/*.gz | wc -l
        10000
        $ zcat t2/*.gz | sort -k 1,1n | md5sum
        72d4ff27a28afbc066d5804999d5a504  -
        $ zcat t2.gz | md5sum
        72d4ff27a28afbc066d5804999d5a504  -

## splitxlsx

Usage

```text
split XLSX sheet into multiple sheets according to column values

Strengths: Sheet properties are remained unchanged.
Weakness : Complicated sheet structures are not well supported, e.g.,
  1. merged cells
  2. more than one header row

Usage:
  csvtk splitxlsx [flags]

Flags:
  -f, --fields string       comma separated key fields, column name or index. e.g. -f 1-3 or -f id,id2 or -F -f "group*" (default "1")
  -F, --fuzzy-fields        using fuzzy fields, e.g., -F -f "*name" or -F -f "id123*"
  -h, --help                help for splitxlsx
  -i, --ignore-case         ignore case (cell value)
  -a, --list-sheets         list all sheets
  -N, --sheet-index int     Nth sheet to retrieve (default 1)
  -n, --sheet-name string   sheet to retrieve
```

Examples


1. example data

        # list all sheets
        $ csvtk xlsx2csv -a accounts.xlsx
        index   sheet
        1       names
        2       phones
        3       region

        # data of sheet "names"
        $ csvtk xlsx2csv accounts.xlsx | csvtk pretty
        id   first_name   last_name   username
        11   Rob          Pike        rob
        2    Ken          Thompson    ken
        4    Robert       Griesemer   gri
        1    Robert       Thompson    abc
        NA   Robert       Abel        123

1. split sheet "names" according to `first_name`

        $ csvtk splitxlsx accounts.xlsx -n names -f first_name

        $ ls accounts.*
        accounts.split.xlsx  accounts.xlsx

        $ csvtk splitxlsx -a accounts.split.xlsx
        index   sheet
        1       names
        2       phones
        3       region
        4       Rob
        5       Ken
        6       Robert

        $ csvtk xlsx2csv accounts.split.xlsx -n Rob \
            | csvtk pretty
        id   first_name   last_name   username
        11   Rob          Pike        rob

        $ csvtk xlsx2csv accounts.split.xlsx -n Robert \
            | csvtk pretty
        id   first_name   last_name   username
        4    Robert       Griesemer   gri
        1    Robert       Thompson    abc
        NA   Robert       Abel        123

## collapse

Usage

```text
collapse one field with selected fields as keys

Usage:
  csvtk collapse [flags]

Flags:
  -f, --fields string      key fields. e.g -f 1,2 or -f columnA,columnB (default "1")
  -F, --fuzzy-fields       using fuzzy fields (only for key fields), e.g., -F -f "*name" or -F -f "id123*"
  -h, --help               help for collapse
  -i, --ignore-case        ignore case
  -s, --separater string   separater for collapsed data (default "; ")
  -v, --vfield string      value field

```

examples

1. data

        $ csvtk pretty teachers.csv
        lab                     teacher   class
        computational biology   Tom       Bioinformatics
        computational biology   Tom       Statistics
        computational biology   Rob       Bioinformatics
        sequencing center       Jerry     Bioinformatics
        sequencing center       Nick      Molecular Biology
        sequencing center       Nick      Microbiology

1. List teachers for every lab/class. `uniq` is used to deduplicate items.

        $ cat teachers.csv  \
            | csvtk uniq -f lab,teacher  \
            | csvtk collapse -f lab -v teacher \
            | csvtk pretty

        lab                     teacher
        computational biology   Tom; Rob
        sequencing center       Jerry; Nick

        $ cat teachers.csv  \
            | csvtk uniq -f class,teacher  \
            | csvtk collapse -f class -v teacher -s ", " \
            | csvtk pretty

        class               teacher
        Statistics          Tom
        Bioinformatics      Tom, Rob, Jerry
        Molecular Biology   Nick
        Microbiology        Nick

1. Multiple key fields supported

        $ cat teachers.csv  \
            | csvtk collapse -f teacher,lab -v class \
            | csvtk pretty

        teacher   lab                     class
        Tom       computational biology   Bioinformatics; Statistics
        Rob       computational biology   Bioinformatics
        Jerry     sequencing center       Bioinformatics
        Nick      sequencing center       Molecular Biology; Microbiology

## comb

Usage

```
compute combinations of items at every row

Usage:
  csvtk comb [flags]

Aliases:
  comb, combination

Flags:
  -h, --help          help for comb
  -i, --ignore-case   ignore-case
  -S, --nat-sort      sort items in natural order
  -n, --number int    number of items in a combination, 0 for no limit, i.e., return all combinations (default 2)
  -s, --sort          sort items in a combination
```

Examples:

```shell
$ cat players.csv 
gender,id,name
male,1,A
male,2,B
male,3,C
female,11,a
female,12,b
female,13,c
female,14,d

# put names of one group in one row
$ cat players.csv | csvtk collapse -f 1 -v 3 -s ';' | csvtk cut -f 2 
name
A;B;C
a;b;c;d

# n = 2
$ cat players.csv | csvtk collapse -f 1 -v 3 -s ';' | csvtk cut -f 2 | csvtk comb -d ';' -n 2
A,B
A,C
B,C
a,b
a,c
b,c
a,d
b,d
c,d

# n = 3
$ cat players.csv | csvtk collapse -f 1 -v 3 -s ';' | csvtk cut -f 2 | csvtk comb -d ';' -n 3
A,B,C
a,b,c
a,b,d
a,c,d
b,c,d

# n = 0
$ cat players.csv | csvtk collapse -f 1 -v 3 -s ';' | csvtk cut -f 2 | csvtk comb -d ';' -n 0
A
B
A,B
C
A,C
B,C
A,B,C
a
b
a,b
c
a,c
b,c
a,b,c
d
a,d
b,d
a,b,d
c,d
a,c,d
b,c,d
a,b,c,d

```
        
## add-header

Usage

```text
add column names

Usage:
  csvtk add-header [flags]

Flags:
  -h, --help            help for add-header
  -n, --names strings   column names to add, in CSV format

```

Examples:

1. No new colnames given:

        $ seq 3 | csvtk mutate -H \
            | csvtk add-header
        [WARN] colnames not given, c1, c2, c3... will be used
        c1,c2
        1,1
        2,2
        3,3

1. Adding new colnames:

        $ seq 3 | csvtk mutate -H \
            | csvtk add-header -n a,b
        a,b
        1,1
        2,2
        3,3
        $ seq 3 | csvtk mutate -H \
            | csvtk add-header -n a -n b
        a,b
        1,1
        2,2
        3,3

        $ seq 3 | csvtk mutate -H -t \
            | csvtk add-header -t -n a,b
        a       b
        1       1
        2       2
        3       3

## del-header

Usage

```text
delete column names

Usage:
  csvtk del-header [flags]

Flags:
  -h, --help   help for del-header

```

Examples:

    $ seq 3 | csvtk add-header
    c1
    1
    2
    3

    $ seq 3 | csvtk add-header | csvtk del-header
    1
    2
    3

    $ seq 3 | csvtk del-header -H
    1
    2
    3

## rename

Usage

```text
rename column names with new names

Usage:
  csvtk rename [flags]

Flags:
  -f, --fields string   select only these fields. e.g -f 1,2 or -f columnA,columnB
  -F, --fuzzy-fields    using fuzzy fields, e.g., -F -f "*name" or -F -f "id123*"
  -n, --names string    comma separated new names

```

Examples:

- Setting new names: `csvtk rename -f A,B -n a,b` or `csvtk rename -f 1-3 -n a,b,c`

        $ cat testdata/phones.csv
        username,phone
        gri,11111
        rob,12345
        ken,22222
        shenwei,999999

        $ cat testdata/phones.csv \
            | csvtk rename -f 1-2 -n 姓名,电话 \
            | csvtk pretty 
        姓名      电话
        gri       11111
        rob       12345
        ken       22222
        shenwei   999999

## rename2

Usage

```text
rename column names by regular expression

Special replacement symbols:

    {nr}  ascending number, starting from 1
    {kv}  Corresponding value of the key (captured variable $n) by key-value file,
          n can be specified by flag --key-capt-idx (default: 1)

Usage:
  csvtk rename2 [flags]

Flags:
  -f, --fields string          select only these fields. e.g -f 1,2 or -f columnA,columnB
  -F, --fuzzy-fields           using fuzzy fields, e.g., -F -f "*name" or -F -f "id123*"
  -h, --help                   help for rename2
  -i, --ignore-case            ignore case
  -K, --keep-key               keep the key as value when no value found for the key
      --key-capt-idx int       capture variable index of key (1-based) (default 1)
      --key-miss-repl string   replacement for key with no corresponding value
  -k, --kv-file string         tab-delimited key-value file for replacing key with value when using "{kv}" in -r (--replacement)
  -p, --pattern string         search regular expression
  -r, --replacement string     renamement. supporting capture variables.  e.g. $1 represents the text of the first submatch. ATTENTION: use SINGLE quote NOT double quotes in *nix OS or use the \ escape character. Ascending number is also supported by "{nr}".use ${1} instead of $1 when {kv} given!

```

Examples:

- Add suffix to all column names.

        $ cat testdata/phones.csv
        username,phone
        gri,11111
        rob,12345
        ken,22222
        shenwei,999999

        $ cat testdata/phones.csv \
            | csvtk rename2 -F -f "*" -p "(.*)" -r 'prefix_${1}_suffix'
        prefix_username_suffix,prefix_phone_suffix
        gri,11111
        rob,12345
        ken,22222
        shenwei,999999

- supporting `{kv}` and `{nr}` in `csvtk replace`. e.g., replace barcode with sample name.

        $ cat barcodes.tsv
        Sample  Barcode
        sc1     CCTAGATTAAT
        sc2     GAAGACTTGGT
        sc3     GAAGCAGTATG
        sc4     GGTAACCTGAC
        sc5     ATAGTTCTCGT

        $ cat table.tsv
        gene    ATAGTTCTCGT     GAAGCAGTATG     GAAGACTTGGT     AAAAAAAAAA
        gene1   0       0       3       0
        gen1e2  0       0       0       0

        # note that, we must arrange the order of barcodes.tsv to KEY-VALUE
        $ csvtk cut -t -f 2,1 barcodes.tsv
        Barcode Sample
        CCTAGATTAAT     sc1
        GAAGACTTGGT     sc2
        GAAGCAGTATG     sc3
        GGTAACCTGAC     sc4
        ATAGTTCTCGT     sc5

        # here we go!!!!

        $ csvtk rename2 -t -k <(csvtk cut -t -f 2,1 barcodes.tsv) \
            -f -1 -p '(.+)' -r '{kv}' --key-miss-repl unknown table.tsv
        gene    sc5     sc3     sc2     unknown
        gene1   0       0       3       0
        gen1e2  0       0       0       0

- `{nr}`, incase you need this

        $ echo "a,b,c,d" \
            | csvtk rename2  -p '(.+)' -r 'col_{nr}' -f -1 --start-num 2
        a,col_2,col_3,col_4

## replace

Usage

```text
replace data of selected fields by regular expression

Note that the replacement supports capture variables.
e.g. $1 represents the text of the first submatch.
ATTENTION: use SINGLE quote NOT double quotes in *nix OS.

Examples: Adding space to all bases.

  csvtk replace -p "(.)" -r '$1 ' -s

Or use the \ escape character.

  csvtk replace -p "(.)" -r "\$1 " -s

more on: http://shenwei356.github.io/csvtk/usage/#replace

Special replacement symbols:

  {nr}    Record number, starting from 1
  {kv}    Corresponding value of the key (captured variable $n) by key-value file,
          n can be specified by flag --key-capt-idx (default: 1)

Usage:
  csvtk replace [flags]


Flags:
  -f, --fields string          select only these fields. e.g -f 1,2 or -f columnA,columnB (default "1")
  -F, --fuzzy-fields           using fuzzy fields, e.g., -F -f "*name" or -F -f "id123*"
  -i, --ignore-case            ignore case
  -K, --keep-key               keep the key as value when no value found for the key
      --key-capt-idx int       capture variable index of key (1-based) (default 1)
      --key-miss-repl string   replacement for key with no corresponding value
  -k, --kv-file string         tab-delimited key-value file for replacing key with value when using "{kv}" in -r (--replacement)
  -p, --pattern string         search regular expression
  -r, --replacement string     replacement. supporting capture variables.  e.g. $1 represents the text of the first submatch. ATTENTION: for *nix OS, use SINGLE quote NOT double quotes or use the \ escape character. Record number is also supported by "{nr}".use ${1} instead of $1 when {kv} given!

```

Examples

- remove Chinese charactors

        $ csvtk replace -F -f "*_name" -p "\p{Han}+" -r ""
        
- replace by key-value files

        $ cat data.tsv
        name    id
        A       ID001
        B       ID002
        C       ID004

        $ cat alias.tsv
        001     Tom
        002     Bob
        003     Jim

        $ csvtk replace -t -f 2  -p "ID(.+)" -r "N: {nr}, alias: {kv}" -k alias.tsv  data.tsv
        [INFO] read key-value file: alias.tsv
        [INFO] 3 pairs of key-value loaded
        name    id
        A       N: 1, alias: Tom
        B       N: 2, alias: Bob
        C       N: 3, alias: 004

## mutate

Usage

```text
create new column from selected fields by regular expression

Usage:
  csvtk mutate [flags]

Flags:
  -f, --fields string    select only these fields. e.g -f 1,2 or -f columnA,columnB (default "1")
  -i, --ignore-case      ignore case
      --na               for unmatched data, use blank instead of original data
  -n, --name string      new column name
  -p, --pattern string   search regular expression with capture bracket. e.g. (default "^(.+)$")

```

Examples

- By default, copy a column: `csvtk mutate -f id -n newname`

- Extract prefix of data as group name using regular expression (get "A" from "A.1" as group name):

        csvtk mutate -f sample -n group -p "^(.+?)\."
        
- get the first letter as new column

        $ cat testdata/phones.csv
        username,phone
        gri,11111
        rob,12345
        ken,22222
        shenwei,999999

        $ cat testdata/phones.csv \ 
            | csvtk mutate -f username -p "^(\w)" -n first_letter
        username,phone,first_letter
        gri,11111,g
        rob,12345,r
        ken,22222,k
        shenwei,999999,s

## mutate2

Usage

```text
create new column from selected fields by awk-like artithmetic/string expressions

The artithmetic/string expression is supported by:

  https://github.com/Knetic/govaluate

Supported operators and types:

  Modifiers: + - / * & | ^ ** % >> <<
  Comparators: > >= < <= == != =~ !~
  Logical ops: || &&
  Numeric constants, as 64-bit floating point (12345.678)
  String constants (single quotes: 'foobar')
  Date constants (single quotes)
  Boolean constants: true false
  Parenthesis to control order of evaluation ( )
  Arrays (anything separated by , within parenthesis: (1, 2, 'foo'))
  Prefixes: ! - ~
  Ternary conditional: ? :
  Null coalescence: ??

Usage:
  csvtk mutate2 [flags]

Flags:
  -L, --digits int          number of digits after the dot (default 2)
  -s, --digits-as-string    treate digits as string to avoid converting big digits into scientific notation
  -e, --expression string   arithmetic/string expressions. e.g. "'string'", '"abc"', ' $a + "-" + $b ', '$1 + $2', '$a / $b', ' $1 > 100 ? "big" : "small" '
  -h, --help                help for mutate2
  -n, --name string         new column name

```

Example

1. Constants

        $ cat testdata/digitals.tsv \
            | csvtk mutate2 -t -H -e " 'abc' "
        4       5       6       abc
        1       2       3       abc
        7       8       0       abc
        8       1,000   4       abc

        $ val=123 \
            && cat testdata/digitals.tsv \
            | csvtk mutate2 -t -H -e " $val "
        4       5       6       123
        1       2       3       123
        7       8       0       123
        8       1,000   4       123

1. String concatenation

        $ cat testdata/names.csv  \
            | csvtk mutate2 -n full_name -e ' $first_name + " " + $last_name ' \
            | csvtk pretty
        id   first_name   last_name   username   full_name
        11   Rob          Pike        rob        Rob Pike
        2    Ken          Thompson    ken        Ken Thompson
        4    Robert       Griesemer   gri        Robert Griesemer
        1    Robert       Thompson    abc        Robert Thompson
        NA   Robert       Abel        123        Robert Abel

1. Math

        $ cat testdata/digitals.tsv | csvtk mutate2 -t -H -e '$1 + $3' -L 0
        4       5       6       10
        1       2       3       4
        7       8       0       7
        8       1,000   4       12

1. Bool

        $ cat testdata/digitals.tsv | csvtk mutate2 -t -H -e '$1 > 5'
        4       5       6       false
        1       2       3       false
        7       8       0       true
        8       1,000   4       true

1. Ternary conditional

        $ cat testdata/digitals.tsv | csvtk mutate2 -t -H -e '$1 > 5 ? "big" : "small" '
        4       5       6       small
        1       2       3       small
        7       8       0       big
        8       1,000   4       big

## sep

Usage

```text
separate column into multiple columns

Usage:
  csvtk sep [flags]

Flags:
      --drop            drop extra data, exclusive with --merge
  -f, --fields string   select only these fields. e.g -f 1,2 or -f columnA,columnB (default "1")
  -h, --help            help for sep
  -i, --ignore-case     ignore case
      --merge           only splits at most N times, exclusive with --drop
      --na string       content for filling NA data
  -n, --names strings   new column names
  -N, --num-cols int    preset number of new created columns
  -R, --remove          remove input column
  -s, --sep string      separator
  -r, --use-regexp      separator is a regular expression
```

Examples:

```shell
$ cat players.csv | csvtk collapse -f 1 -v 3 -s ';'
gender,name
male,A;B;C
female,a;b;c;d

# set number of new columns as 3.
$ cat players.csv | csvtk collapse -f 1 -v 3 -s ';' \
    | csvtk sep -f 2 -s ';' -n p1,p2,p3,p4 -N 4 --na NA \
    | csvtk pretty
gender   name      p1   p2   p3   p4
male     A;B;C     A    B    C    NA
female   a;b;c;d   a    b    c    d
    
# set number of new columns as 3, drop extra values 
$ cat players.csv | csvtk collapse -f 1 -v 3 -s ';' \
    | csvtk sep -f 2 -s ';' -n p1,p2,p3  --drop \
    | csvtk pretty
gender   name      p1   p2   p3
male     A;B;C     A    B    C
female   a;b;c;d   a    b    c

# set number of new columns as 3, split as most 3 parts
$ cat players.csv | csvtk collapse -f 1 -v 3 -s ';' \
    | csvtk sep -f 2 -s ';' -n p1,p2,p3  --merge \
    | csvtk pretty
gender   name      p1   p2   p3
male     A;B;C     A    B    C
female   a;b;c;d   a    b    c;d

#
$ echo -ne "taxid\tlineage\n9606\tEukaryota;Chordata;Mammalia;Primates;Hominidae;Homo;Homo sapiens\n"
taxid   lineage
9606    Eukaryota;Chordata;Mammalia;Primates;Hominidae;Homo;Homo sapiens

$ echo -ne "taxid\tlineage\n9606\tEukaryota;Chordata;Mammalia;Primates;Hominidae;Homo;Homo sapiens\n" \
    | csvtk sep -t -f 2 -s ';' -n kindom,phylum,class,order,family,genus,species --remove \
    | csvtk pretty -t
taxid   kindom      phylum     class      order      family      genus   species
9606    Eukaryota   Chordata   Mammalia   Primates   Hominidae   Homo    Homo sapiens
```
## gather

Usage

```text
gather columns into key-value pairs

Usage:
  csvtk gather [flags]

Flags:
  -f, --fields string   fields for gathering. e.g -f 1,2 or -f columnA,columnB, or -f -columnA for unselect columnA
  -F, --fuzzy-fields    using fuzzy fields, e.g., -F -f "*name" or -F -f "id123*"
  -k, --key string      name of key column to create in output
  -v, --value string    name of value column to create in output

```

Examples:

    $ cat testdata/names.csv
    id,first_name,last_name,username
    11,"Rob","Pike",rob
    2,Ken,Thompson,ken
    4,"Robert","Griesemer","gri"
    1,"Robert","Thompson","abc"
    NA,"Robert","Abel","123

    $ cat testdata/names.csv \
        | csvtk gather -k item -v value -f -1
    id,item,value
    11,first_name,Rob
    11,last_name,Pike
    11,username,rob
    2,first_name,Ken
    2,last_name,Thompson
    2,username,ken
    4,first_name,Robert
    4,last_name,Griesemer
    4,username,gri
    1,first_name,Robert
    1,last_name,Thompson
    1,username,abc
    NA,first_name,Robert
    NA,last_name,Abel
    NA,username,123

## sort

Usage

```text
sort by selected fields

Usage:
  csvtk sort [flags]

Flags:
  -h, --help             help for sort
  -i, --ignore-case      ignore-case
  -k, --keys strings     keys (multiple values supported). sort type supported, "N" for natural order, "n" for number, "u" for user-defined order and "r" for reverse. e.g., "-k 1" or "-k A:r" or ""-k 1:nr -k 2" (default [1])
  -L, --levels strings   user-defined level file (one level per line, multiple values supported). format: <field>:<level-file>.  e.g., "-k name:u -L name:level.txt"
```

Examples

- data

        $ cat testdata/names.csv
        id,first_name,last_name,username
        11,"Rob","Pike",rob
        2,Ken,Thompson,ken
        4,"Robert","Griesemer","gri"
        1,"Robert","Thompson","abc"
        NA,"Robert","Abel","123"

- By single column : `csvtk sort -k 1` or `csvtk sort -k last_name`

    - in alphabetical order

            $ cat testdata/names.csv \
                | csvtk sort -k first_name
            id,first_name,last_name,username
            2,Ken,Thompson,ken
            11,Rob,Pike,rob
            NA,Robert,Abel,123
            1,Robert,Thompson,abc
            4,Robert,Griesemer,gri

    - in reversed alphabetical order (`key:r`)

            $ cat testdata/names.csv \
                | csvtk sort -k first_name:r
            id,first_name,last_name,username
            NA,Robert,Abel,123
            1,Robert,Thompson,abc
            4,Robert,Griesemer,gri
            11,Rob,Pike,rob
            2,Ken,Thompson,ken

    - in numerical order (`key:n`)

            $ cat testdata/names.csv \
                | csvtk sort -k id:n
            id,first_name,last_name,username
            NA,Robert,Abel,123
            1,Robert,Thompson,abc
            2,Ken,Thompson,ken
            4,Robert,Griesemer,gri
            11,Rob,Pike,rob

    - in natural order (`key:N`)

            $ cat testdata/names.csv | csvtk sort -k id:N
            id,first_name,last_name,username
            1,Robert,Thompson,abc
            2,Ken,Thompson,ken
            4,Robert,Griesemer,gri
            11,Rob,Pike,rob
            NA,Robert,Abel,123
            
    - in natural order (`key:N`), a bioinformatics example
    
            $ echo "X,Y,1,10,2,M,11,1_c,Un_g,1_g" | csvtk transpose 
            X
            Y
            1
            10
            2
            M
            11
            1_c
            Un_g
            1_g

            $ echo "X,Y,1,10,2,M,11,1_c,Un_g,1_g" \
                | csvtk transpose \
                | csvtk sort -H -k 1:N
            1
            1_c
            1_g
            2
            10
            11
            M
            Un_g
            X
            Y

- By multiple columns: `csvtk sort -k 1,2` or `csvtk sort -k 1 -k 2` or `csvtk sort -k last_name,age`

        # by first_name and then last_name
        $ cat testdata/names.csv | csvtk sort -k first_name -k last_name
        id,first_name,last_name,username
        2,Ken,Thompson,ken
        11,Rob,Pike,rob
        NA,Robert,Abel,123
        4,Robert,Griesemer,gri
        1,Robert,Thompson,abc

        # by first_name and then ID
        $ cat testdata/names.csv | csvtk sort -k first_name -k id:n
        id,first_name,last_name,username
        2,Ken,Thompson,ken
        11,Rob,Pike,rob
        NA,Robert,Abel,123
        1,Robert,Thompson,abc
        4,Robert,Griesemer,gri

- By ***user-defined order***

        # user-defined order/level
        $ cat testdata/size_level.txt
        tiny
        mini
        small
        medium
        big

        # original data
        $ cat testdata/size.csv
        id,size
        1,Huge
        2,Tiny
        3,Big
        4,Small
        5,Medium

        $ csvtk sort -k 2:u -i -L 2:testdata/size_level.txt testdata/size.csv
        id,size
        2,Tiny
        4,Small
        5,Medium
        3,Big
        1,Huge

## plot

Usage

```text
plot common figures

Notes:

  1. Output file can be set by flag -o/--out-file.
  2. File format is determined by the out file suffix.
     Supported formats: eps, jpg|jpeg, pdf, png, svg, and tif|tiff
  3. If flag -o/--out-file not set (default), image is written to stdout,
     you can display the image by pipping to "display" command of Imagemagic
     or just redirect to file.

Usage:
  csvtk plot [command]

Available Commands:
  box         plot boxplot
  hist        plot histogram
  line        line plot and scatter plot

Flags:
      --axis-width float     axis width (default 1.5)
  -f, --data-field string    column index or column name of data (default "1")
      --format string        image format for stdout when flag -o/--out-file not given. available values: eps, jpg|jpeg, pdf, png, svg, and tif|tiff. (default "png")
  -g, --group-field string   column index or column name of group
      --height float         Figure height (default 4.5)
      --label-size int       label font size (default 14)
      --tick-width float     axis tick width (default 1.5)
      --title string         Figure title
      --title-size int       title font size (default 16)
      --width float          Figure width (default 6)
      --x-max string         maximum value of X axis
      --x-min string         minimum value of X axis
      --xlab string          x label text
      --y-max string         maximum value of Y axis
      --y-min string         minimum value of Y axis
      --ylab string          y label text

```

***Note that most of the flags of `plot` are global flags of the subcommands
`hist`, `box` and `line`***

**Notes of image output**

1. Output file can be set by flag -o/--out-file.
2. File format is determined by the out file suffix.
   Supported formats: eps, jpg|jpeg, pdf, png, svg, and tif|tiff
3. If flag -o/--out-file not set (default), image is written to stdout,
   you can display the image by pipping to  `display` command of `Imagemagic`
   or just redirect to file.

## plot hist

Usage

```text
plot histogram

Notes:

  1. Output file can be set by flag -o/--out-file.
  2. File format is determined by the out file suffix.
     Supported formats: eps, jpg|jpeg, pdf, png, svg, and tif|tiff
  3. If flag -o/--out-file not set (default), image is written to stdout,
     you can display the image by pipping to "display" command of Imagemagic
     or just redirect to file.

Usage:
  csvtk plot hist [flags]

Flags:
      --bins int          number of bins (default 50)
      --color-index int   color index, 1-7 (default 1)

```

Examples

- example data

        $ zcat testdata/grouped_data.tsv.gz | head -n 5 | csvtk -t pretty
        Group     Length   GC Content
        Group A   97       57.73
        Group A   95       49.47
        Group A   97       49.48
        Group A   100      51.00

- plot histogram with data of the second column:

        $ csvtk -t plot hist testdata/grouped_data.tsv.gz -f 2 \
            --title Histogram -o histogram.png

    ![histogram.png](testdata/figures/histogram.png)

- You can also write image to stdout and pipe to "display" command of Imagemagic:
    
        $ csvtk -t plot hist testdata/grouped_data.tsv.gz -f 2 | display


## plot box

Usage

```text
plot boxplot

Notes:

  1. Output file can be set by flag -o/--out-file.
  2. File format is determined by the out file suffix.
     Supported formats: eps, jpg|jpeg, pdf, png, svg, and tif|tiff
  3. If flag -o/--out-file not set (default), image is written to stdout,
     you can display the image by pipping to "display" command of Imagemagic
     or just redirect to file.

Usage:
  csvtk plot box [flags]

Flags:
      --box-width float   box width
      --horiz             horize box plot

```

Examples

- plot boxplot with data of the "GC Content" (third) column,
group information is the "Group" column.

        csvtk -t plot box testdata/grouped_data.tsv.gz -g "Group" -f "GC Content" \
            --width 3 --title "Box plot" \
            > boxplot.png

    ![boxplot.png](testdata/figures/boxplot.png)

- plot horiz boxplot with data of the "Length" (second) column,
group information is the "Group" column.

        $ csvtk -t plot box testdata/grouped_data.tsv.gz -g "Group" -f "Length" \
            --height 3 --width 5 --horiz --title "Horiz box plot" \
            > boxplot2.png`

    ![boxplot2.png](testdata/figures/boxplot2.png)

## plot line

Usage

```text
line plot and scatter plot

Notes:

  1. Output file can be set by flag -o/--out-file.
  2. File format is determined by the out file suffix.
     Supported formats: eps, jpg|jpeg, pdf, png, svg, and tif|tiff
  3. If flag -o/--out-file not set (default), image is written to stdout,
     you can display the image by pipping to "display" command of Imagemagic
     or just redirect to file.

Usage:
  csvtk plot line [flags]

Flags:
  -x, --data-field-x string   column index or column name of X for command line
  -y, --data-field-y string   column index or column name of Y for command line
      --legend-left           locate legend along the left edge of the plot
      --legend-top            locate legend along the top edge of the plot
      --line-width float      line width (default 1.5)
      --point-size float      point size (default 3)
      --scatter               only plot points

```

Examples

- example data

        $ head -n 5 testdata/xy.tsv
        Group   X       Y
        A       0       1
        A       1       1.3
        A       1.5     1.5
        A       2.0     2

- plot line plot with X-Y data

        $ csvtk -t plot line testdata/xy.tsv -x X -y Y -g Group \
            --title "Line plot" \
            > lineplot.png

    ![lineplot.png](testdata/figures/lineplot.png)

- plot scatter

        $ csvtk -t plot line testdata/xy.tsv -x X -y Y -g Group \
            --title "Scatter" --scatter \
            > lineplot.png

    ![scatter.png](testdata/figures/scatter.png)


## cat

Usage

```text
stream file to stdout and report progress on stderr

Usage:
  csvtk cat [flags]

Flags:
  -b, --buffsize int     buffer size (default 8192)
  -h, --help             help for cat
  -L, --lines            count lines instead of bytes
  -p, --print-freq int   print frequency (-1 for print after parsing) (default 1)
  -s, --total int        expected total bytes/lines (default -1)
```

Examples

1. Stream file, report progress in bytes

        csvtk cat file.tsv

2. Stream file from stdin, report progress in lines

        tac input.tsv | csvtk cat -L -s `wc -l < input.tsv` -

## genautocomplete

Usage

```text
generate shell autocompletion script

Note: The current version supports Bash only.
This should work for *nix systems with Bash installed.

Howto:

1. run: csvtk genautocomplete

2. create and edit ~/.bash_completion file if you don't have it.

        nano ~/.bash_completion

   add the following:

        for bcfile in ~/.bash_completion.d/* ; do
          . $bcfile
        done

Usage:
  csvtk genautocomplete [flags]

Flags:
      --file string   autocompletion file (default "/home/shenwei/.bash_completion.d/csvtk.sh")
  -h, --help          help for genautocomplete
      --type string   autocompletion type (currently only bash supported) (default "bash")

```

<div id="disqus_thread"></div>
<script>
/**
* RECOMMENDED CONFIGURATION VARIABLES: EDIT AND UNCOMMENT THE SECTION BELOW TO INSERT DYNAMIC VALUES FROM YOUR PLATFORM OR CMS.
* LEARN WHY DEFINING THESE VARIABLES IS IMPORTANT: https://disqus.com/admin/universalcode/#configuration-variables
*/
/*
var disqus_config = function () {
this.page.url = PAGE_URL; // Replace PAGE_URL with your page's canonical URL variable
this.page.identifier = PAGE_IDENTIFIER; // Replace PAGE_IDENTIFIER with your page's unique identifier variable
};
*/
(function() { // DON'T EDIT BELOW THIS LINE
var d = document, s = d.createElement('script');

s.src = '//csvtk.disqus.com/embed.js';

s.setAttribute('data-timestamp', +new Date());
(d.head || d.body).appendChild(s);
})();
</script>
<noscript>Please enable JavaScript to view the <a href="https://disqus.com/?ref_noscript" rel="nofollow">comments powered by Disqus.</a></noscript>
