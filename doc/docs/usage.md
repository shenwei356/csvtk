# Usage and Examples

## Before use

**Attention**

1. The CSV parser requires all the lines have same number of fields/columns.
 Even lines with spaces will cause error.
2. By default, csvtk think your files have header row, if not, use `-H`.
3. By default, lines starting with `#` will be ignored, if the header row
 starts with `#`, please assign `-C` another rare symbol, e.g. `&`.
4. By default, csvtk handles CSV files, use `-t` for tab-delimited files.

## csvkit

Usage

```
Another cross-platform, efficient and practical CSV/TSV toolkit

Version: 0.2.4

Author: Wei Shen <shenwei356@gmail.com>

Documents  : http://shenwei356.github.io/csvtk
Source code: https://github.com/shenwei356/csvtk

Attention:

  1. The CSV parser requires all the lines have same number of fields/columns.
     Even lines with spaces will cause error.
  2. By default, csvtk think your files have header row, if not, use "-H".
  3. By default, lines starting with '#' will be ignored, if the header row
     starts with '#', please assign "-C" another rare symbol, e.g. '&'.
  4. By default, csvtk handles CSV files, use "-t" for tab-delimited files.

Usage:
  csvtk [command]

Available Commands:
  csv2tab     convert CSV to tabular format
  cut         select parts of fields
  grep        grep data by selected fields with patterns/regular expressions
  inter       intersection of multiple files
  join        join multiple CSV files by selected fields
  mutate      create new column from selected fields by regular expression
  rename      rename column names
  rename2     rename column names by regular expression
  replace     replace data of selected fields by regular expression
  sort        sort by selected fields
  space2tab   convert space delimited format to CSV
  stat        summary of CSV file
  tab2csv     convert tabular format to CSV
  transpose   transpose CSV data
  uniq        unique data without sorting

Flags:
  -c, --chunk-size int         chunk size of CSV reader (default 50)
  -C, --comment-char string    lines starting with commment-character will be ignored. if your header row starts with '#', please assign "-C" another rare symbol, e.g. '&' (default "#")
  -d, --delimiter string       delimiting character of the input CSV file (default ",")
  -l, --lazy-quotes            if given, a quote may appear in an unquoted field and a non-doubled quote may appear in a quoted field
  -H, --no-header-row          specifies that the input CSV file does not have header row
  -j, --num-cpus int           number of CPUs to use (default value depends on your computer) (default 4)
  -D, --out-delimiter string   delimiting character of the input CSV file (default ",")
  -o, --out-file string        out file ("-" for stdout, suffix .gz for gzipped out) (default "-")
  -T, --out-tabs               specifies that the output is delimited with tabs. Overrides "-D"
  -t, --tabs                   specifies that the input CSV file is delimited with tabs. Overrides "-d"

Use "csvtk [command] --help" for more information about a command.

```

## stat

## csv2tab

## tab2csv

## space2tab

## transpose

Usage

```

```

Examples

    $ cat digitals.tsv
    4       5       6                                                                                  
    1       2       3                                                                                  
    7       8       0
    8       1,000   4

    $ csvtk transpose -t digitals.tsv
    4       1       7       8
    5       2       8       1,000
    6       3       0       4


## cut

Usage

```
elect parts of fields

Usage:
  csvtk cut [flags]

Flags:
  -n, --colnames        print column names
  -f, --fields string   select only these fields. e.g -f 1,2 or -f columnA,columnB
  -F, --fuzzy-fields    using fuzzy fields, e.g. *name or id123*

```

Examples

- Print colnames: `csvtk cut -n`
- By index: `csvtk cut -f 1,2`
- By names: `csvtk cut -f first_name,username`
- **Unselect**: `csvtk cut -f -1,-2` or `csvtk cut -f -first_name`
- **Fuzzy fields**: `csvtk cut -F -f "*_name,username"`
- Field ranges: `csvtk cut -f 2-4` for column 2,3,4 or `csvtk cut -f -3--1` for discarding column 1,2,3
- All fields: `csvtk cut -F -f "*"`

## uniq

Usage

```
unique data without sorting

Usage:
  csvtk uniq [flags]

Flags:
  -f, --fields string   select only these fields. e.g -f 1,2 or -f columnA,columnB (default "1")
  -F, --fuzzy-fields    using fuzzy fileds, e.g. *name or id123*
  -i, --ignore-case     ignore case

```

## inter

Usage

```
intersection of multiple files

Usage:
  csvtk inter [flags]

Flags:
  -f, --fields string   select only these fields. e.g -f 1,2 or -f columnA,columnB (default "1")
  -F, --fuzzy-fields    using fuzzy fileds, e.g. *name or id123*
  -i, --ignore-case     ignore case

```

## grep

Usage

```
grep data by selected fields with patterns/regular expressions

Usage:
  csvtk grep [flags]

Flags:
  -f, --fields string         key field, column name or index (default "1")
  -i, --ignore-case           ignore case
  -v, --invert                invert match
  -p, --pattern value         query pattern (multiple values supported) (default [])
  -P, --pattern-file string   pattern files (could also be CSV format)
  -r, --use-regexp            patterns are regular expression

```

Examples

- By regular expression: `csvtk grep -f first_name -r -p Rob`
- By pattern list: `csvtk grep -f first_name -P name_list.txt`

## join

Usage

```
join 2nd and later files to the first file by selected fields.

Multiple keys supported, but the orders are ignored.

Usage:
  csvtk join [flags]

Flags:
  -f, --fields string    Semicolon seperated key fields of all files, if given one, we think all the files have the same key columns. e.g -f 1;2 or -f A,B;C,D or -f id (default "1")
  -F, --fuzzy-fields     using fuzzy fileds, e.g. *name or id123*
  -i, --ignore-case      ignore case
  -k, --keep-unmatched   keep unmatched data of the first file

```

Examples:

- All files have same key column: `csvtk join -f id file1.csv file2.csv`
- Files have different key columns: `csvtk join -f "username;username;name" names.csv phone.csv adress.csv -k`

## rename

Usage

```
rename column names

Usage:
  csvtk rename [flags]

Flags:
  -f, --fields string   select only these fields. e.g -f 1,2 or -f columnA,columnB
  -F, --fuzzy-fields    using fuzzy fileds, e.g. *name or id123*
  -n, --names string    comma separated new names

```

Examples:

- Setting new names: `csvtk rename -f A,B -n a,b` or `csvtk rename -f 1-3 -n a,b,c`

## rename2

Usage

```
rename column names by regular expression

Usage:
  csvtk rename2 [flags]

Flags:
  -f, --fields string        select only these fields. e.g -f 1,2 or -f columnA,columnB
  -F, --fuzzy-fields         using fuzzy fileds, e.g. *name or id123*
  -i, --ignore-case          ignore case
  -p, --pattern string       search regular expression
  -r, --replacement string   renamement. supporting capture variables.  e.g. $1 represents the text of the first submatch. ATTENTION: use SINGLE quote NOT double quotes in *nix OS or use the \ escape character.

```

Examples:

- replacing with original names by regular express: `cat ../testdata/c.csv | ./csvtk rename2 -F -f "*" -p "(.*)" -r 'prefix_$1'` for adding prefix to all column names.

## replace

Usage

```
replace data of selected fields by regular expression

Usage:
  csvtk replace [flags]

Flags:
  -f, --fields string        select only these fields. e.g -f 1,2 or -f columnA,columnB (default "1")
  -F, --fuzzy-fields         using fuzzy fileds, e.g. *name or id123*
  -i, --ignore-case          ignore case
  -p, --pattern string       search regular expression
  -r, --replacement string   replacement. supporting capture variables.  e.g. $1 represents the text of the first submatch. ATTENTION: use SINGLE quote NOT double quotes in *nix OS or use the \ escape character.

```

Examples

- remove Chinese charactors:  `csvtk replace -F -f "*_name" -p "\p{Han}+" -r ""`

## mutate

Usage

```
create new column from selected fields by regular expression

Usage:
  csvtk mutate [flags]

Flags:
  -f, --fields string    select only these fields. e.g -f 1,2 or -f columnA,columnB (default "1")
  -i, --ignore-case      ignore case
      --na               for unmatched data, use blank instead of orginal data
  -n, --name string      new column name
  -p, --pattern string   search regular expression with capture bracket. e.g. (default "^(.+)$")

```

Examples

- In default, copy a column: `csvtk mutate -f id `
- extract prefix of data as group name (get "A" from "A.1" as group name):
  `csvtk mutate -f sample -n group -p "^(.+?)\."`

## sort

Usage

```
sort by selected fields

Usage:
  csvtk sort [flags]

Flags:
  -k, --keys value   keys. sort type supported, "n" for number and "r" for reverse. e.g. "-k 1" or "-k A:r" or ""-k 1:nr -k 2" (default [1])

```

Examples

- By single column : `csvtk sort -k 1` or `csvtk sort -k last_name`
- By multiple columns: `csvtk sort -k 1,2` or `csvtk sort -k 1 -k 2` or `csvtk sort -k last_name,age`
- Sort by number: `csvtk sort -k 1:n` or  `csvtk sort -k 1:nr` for reverse number
- Complex sort: `csvtk sort -k region -k age:n -k id:nr`



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
