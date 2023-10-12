- [csvtk v0.28.1](https://github.com/shenwei356/csvtk/releases/tag/v0.28.1)
[![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.28.1/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.28.1)
  - `csvtk sort`:
      - support column name containing colons. [#254](https://github.com/shenwei356/csvtk/issues/254)
  - `csvtk filter2:
      - update doc: add the `in` keyword. [#195](https://github.com/shenwei356/csvtk/pull/195)
- [csvtk v0.28.0](https://github.com/shenwei356/csvtk/releases/tag/v0.28.0)
[![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.28.0/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.28.0)
  - `csvtk`:
      - add the shortcut `-X` for the flag `--infile-list`. [#249](https://github.com/shenwei356/csvtk/issues/249)
  - `csvtk pretty`:
      - support field ranges for `-m/--align-center` and `-r/--align-right`. [#244](https://github.com/shenwei356/csvtk/issues/244)
  - `csvtk spread`:
      - support values sharing the same keys. [#248](https://github.com/shenwei356/csvtk/issues/248)
  - `csvtk join`:
      - a new flag `-P/--prefix-duplicates`: add filenames as colname prefixes only for duplicated colnames. [#246](https://github.com/shenwei356/csvtk/issues/246)
  - `csvtk mutate2`:
      - fix changing the order of the header row, the code was accidentally missing during code refactoring in v0.27.0. [#252](https://github.com/shenwei356/csvtk/issues/252)
  - `csvtk xlsx2csv`:
      - fix `open /tmp/excelize-: no such file or directory` error for big `.xlsx` files. [#251](https://github.com/shenwei356/csvtk/issues/251)
  - `csvtk comb`:
      - fix the empty result bug for alphabet sizes greater than 64.
- [csvtk v0.27.2](https://github.com/shenwei356/csvtk/releases/tag/v0.27.2)
[![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.27.2/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.27.2)
  - `csvtk pretty`:
      - fix the bug of empty first row with `-H/--no-header-row`, introduced in v0.27.0.
      - new style `3line` for three-line table.
  - `csvtk csv2xlsx`:
      - binaries compiled with go1.21 would result in a broken xlsx file. [#243](https://github.com/shenwei356/csvtk/issues/243)
  - `csvtk splitxlsx`:
      - fix the error of `invalid worksheet index`. [#1617](https://github.com/qax-os/excelize/issues/1617)
- [csvtk v0.27.1](https://github.com/shenwei356/csvtk/releases/tag/v0.27.1)
[![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.27.1/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.27.1)
  - `csvtk filter2/mutate2`:
      - fix the bug of selecting with field numbers, introduced in v0.27.0. [#242](https://github.com/shenwei356/csvtk/issues/242)
- [csvtk v0.27.0](https://github.com/shenwei356/csvtk/releases/tag/v0.27.0)
[![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.27.0/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.27.0)
  - `csvtk`:
      - code refactoring and simplifying code, with 16% less code.
      - **most commands support open column range syntax**, e.g.,  `csvtk grep -f 2-`. [#120](https://github.com/shenwei356/csvtk/issues/120)
      - **only selected column names are not allowed to be duplicated in the input data**: box, corr, filter, filter2, fold, freq, gather, historysort, inter, join, line, mutate, mutate2, rename, replace, sep, split, summary, unfold, uniq, watch. Other commands do not have the restriction. [#235](https://github.com/shenwei356/csvtk/issues/235)
      - add a new global flag `-Z/--show-row-number`, supported commands: cut, csv2tab, csv2xlsx, tab2csv, pretty.
      - the colum name of row number changes from "n" to "row":  csv2xlsx, csv2tab, cut, filter, filter2, grep, pretty, sample, tab2csv.
  - **new command**:
      - **`csvtk spread`: spread a key-value pair across multiple columns, like tidyr::spread/pivot_wider**.
       [#91](https://github.com/shenwei356/csvtk/issues/91), [#236](https://github.com/shenwei356/csvtk/issues/236), [#239](https://github.com/shenwei356/csvtk/issues/239)
  - `csvtk mutate/mutate2`:
      - **new flags `--at`, `--before`, `--after` for specifying the position of the new column**. [#193](https://github.com/shenwei356/csvtk/issues/193)
  - `csvtk cut`:
      - fix unselect range error. [#234](https://github.com/shenwei356/csvtk/issues/234)
      - fix `-i/--ignore-case`.
  - `csvtk pretty`:
      - **allow align-center and align-right for specific columns**. [#240](https://github.com/shenwei356/csvtk/issues/240)
  - `csvtk round`:
      - fix bug of failing to round scientific notation with value small than one, e.g., `7.1E-1`.
  - `csvtk summary`:
      - fix duplicated columns.
      - fix result error when multiple stats applied to the same column.
  - `csvtk corr/watch`:
      - rewrite and fix bug, support choosing fields with column names.
- [csvtk v0.26.0](https://github.com/shenwei356/csvtk/releases/tag/v0.26.0)
[![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.26.0/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.26.0)
  - `csvtk`: 
      - **near all commands skip empty files now**. [#204](https://github.com/shenwei356/csvtk/issues/204)
      - the global flag `--infile-list` accepts stdin "-". [#210](https://github.com/shenwei356/csvtk/issues/210)
  - new command `csvtk fix`: **fix CSV/TSV with different numbers of columns in rows**. [#226](https://github.com/shenwei356/csvtk/issues/226)
  - `csvtk pretty`: **rewrite to support wrapping cells**. [#206](https://github.com/shenwei356/csvtk/issues/206) [#209](https://github.com/shenwei356/csvtk/issues/209)  [#228](https://github.com/shenwei356/csvtk/issues/228)
  - `csvtk cut/fmtdate/freq/grep/rename/rename2/replace/round`: allow duplicated column names.
  - `csvtk csv2xlsx`: optionally stores numbers as float. [#217](https://github.com/shenwei356/csvtk/issues/217)
  - `csvtk xlsx2csv`: fix bug where `xlsx2csv` treats small number (padj < 1e-25) as 0. It's solved by updating the excelize package. [#261](https://github.com/shenwei356/csvtk/issues/201)
  - `csvtk join`: a new flag for adding filename as column name prefix. by @tetedange13 [#202](https://github.com/shenwei356/csvtk/issues/202)
  - `csvtk mutate2`: fix wrongly treating strings like `E10` as numbers in scientific notation. [#219](https://github.com/shenwei356/csvtk/issues/219)
  - `csvtk sep`: fix the logic. [#218](https://github.com/shenwei356/csvtk/issues/218)
  - `csvtk space2tab`: fix "bufio.Scanner: token too long". [#231](https://github.com/shenwei356/csvtk/issues/231)
- [csvtk v0.25.0](https://github.com/shenwei356/csvtk/releases/tag/v0.25.0)
[![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.25.0/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.25.0)
    - `csvtk`: report empty files.
    - `csvtk join`: fix loading file with no records.
    - `csvtk filter2/muate2`:
        - support variable format of `${var}` with special charactors including commas, spaces, and parentheses, e.g., `${a,b}`, `${a b}`, or `${a (b)}`. [#186](https://github.com/shenwei356/csvtk/issues/186)
    - `csvtk sort`: fix checking non-existed fileds.
    - `csvtk plot box/hist/line`: new flag `--skip-na` for skipping missing data. [#188](https://github.com/shenwei356/csvtk/issues/188)
    - `csvtk csv2xlsx`: stores number as float. [#192](https://github.com/shenwei356/csvtk/issues/192)
    - `csvtk summary`: new functions `argmin`  and `argmax`. [#181](https://github.com/shenwei356/csvtk/issues/181)
- [csvtk v0.24.0](https://github.com/shenwei356/csvtk/releases/tag/v0.24.0)
[![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.24.0/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.24.0)
    - **Incompatible changes**:
        - `csvtk mutate2/summary`:
          - `mutate2`: remove the option `-L/--digits`.
          - use the same option `-w/--decimal-width` to limit floats to N decimal points.
    - new command `csvtk fmtdate`: format date of selected fields. [#159](https://github.com/shenwei356/csvtk/issues/159)
    - `csvtk grep`: fix bug for searching with `-r -p .`.
    - `csvtk csv2rst`: fix bug for data containing unicode. [#137](https://github.com/shenwei356/csvtk/issues/137)
    - `csvtk filter2`: fix bug for date expression. [#146](https://github.com/shenwei356/csvtk/issues/146)
    - `csvtk mutate2/filter2`: 
        - change the way of rexpression evaluation.
        - add custom functions: `len()`. [#153](https://github.com/shenwei356/csvtk/issues/153)
        - **fix bug when using two or more columns with common prefixes in column names**. [#173](https://github.com/shenwei356/csvtk/issues/173)
        - fix value with single or double quotes. [#174](https://github.com/shenwei356/csvtk/issues/174)
    - `csvtk cut`: new flags `-m/--allow-missing-col` and `-b/--blank-missing-col`. [#156](https://github.com/shenwei356/csvtk/issues/156)
    - `csvtk pretty`: still add header row for empty column.
    - `csvtk csv2md`: better format.
    - `csvtk join`: new flag `-n/--ignore-null`. [#163](https://github.com/shenwei356/csvtk/issues/163)
- [csvtk v0.23.0](https://github.com/shenwei356/csvtk/releases/tag/v0.23.0)
[![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.23.0/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.23.0)
    - new comand: `csvtk csv2rst` for converting CSV to reStructuredText format. [#137](https://github.com/shenwei356/csvtk/issues/137)
    - `csvtk pretty`: add header separator line. [#123](https://github.com/shenwei356/csvtk/issues/123)
    - `csvtk mutate2/summary`: fix message and doc. Thanks @VladimirAlexiev  [#127](https://github.com/shenwei356/csvtk/issues/127)
    - `csvtk mutate2`: fix null coalescence: ??. [#129](https://github.com/shenwei356/csvtk/issues/129)
    - `csvtk genautocomplete`: supports bash|zsh|fish|powershell. [#126](https://github.com/shenwei356/csvtk/issues/126)
    - `csvtk cat`: fix progress bar. [#130](https://github.com/shenwei356/csvtk/issues/130)
    - `csvtk grep`: new flag `immediate-output`.
    - `csvtk csv2xlsx`: fix bug for table with > 26 columns. [138](https://github.com/shenwei356/csvtk/issues/138)
- [csvtk v0.22.0](https://github.com/shenwei356/csvtk/releases/tag/v0.22.0)
[![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.22.0/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.22.0)
    - `csvtk`:
        - **global flag `-t` does not overide `-D` anymore**. [#114](https://github.com/shenwei356/csvtk/issues/114)
        - **If the executable/symlink name is `tsvtk` the `-t/--tabs` option for tab input is set**. Thanks @bsipos. [#117](https://github.com/shenwei356/csvtk/pull/117)
    - new command: `csvtk csv2xlsx` for converting CSV/TSV file(s) to a single `.xlsx` file.
    - new command: `csvtk unfold` for unfolding multiple values in cells of a field. [#103](https://github.com/shenwei356/csvtk/issues/103)
    - rename `csvtk collapse` to `csvtk fold`, for folding multiple values of a field into cells of groups.
    - `csvtk cut`: **support range format `2-` to choose 2nd column to the end**. [#106](https://github.com/shenwei356/csvtk/issues/106)
    - `csvtk round`: fix bug of failing to round scientific notation with value small than one, e.g., `7.1E-1`.
- [csvtk v0.21.0](https://github.com/shenwei356/csvtk/releases/tag/v0.21.0)
[![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.21.0/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.21.0)
    - new command: `csvtk nrow/ncol` for printing number of rows or columns.
    - new command: `round` to round float to n decimal places. [#112](https://github.com/shenwei356/csvtk/issues/112)
    - `csvtk headers`: file name and column index is optional outputted with new flag `-v/--verbose`.
    - `csvtk dim`: new flags `--tabluar`, `--cols`, `--rows`, `-n/--no-files`.
    - `csvtk dim/ncol/nrow`: can handle empty files now. [#108](https://github.com/shenwei356/csvtk/issues/108)
    - `csvtk csv2json` [#104](https://github.com/shenwei356/csvtk/issues/104):
        - new flag `-b/--blank`: do not convert "", "na", "n/a", "none", "null", "." to null
        - new flag `-n/--parse-num`: parse numeric values for nth column(s), multiple values are supported and "a"/"all" for all columns.
    - `csvtk xlsx2csv`: fix output for ragged table. [#110](https://github.com/shenwei356/csvtk/issues/110)
    - `csvtk join`: fix bug for joining >2 files.
    - `csvtk uniq`: new flag `-n/--keep-n` for keeping first N records of every key.
    - `csvtk cut`: support repeatedly selecting columns. [#106](https://github.com/shenwei356/csvtk/issues/106)
- [csvtk v0.20.0](https://github.com/shenwei356/csvtk/releases/tag/v0.20.0)
[![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.20.0/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.20.0)
    - new command `csvtk comb`: compute combinations of items at every row.
    - new command `csvtk sep`: separate column into multiple columns. [#96](https://github.com/shenwei356/csvtk/issues/96)
    - `csvtk`:
        - list lines' number of illegal (`-I`) and empty (`-E`) rows. [#97](https://github.com/shenwei356/csvtk/issues/97)
        - new flag `--infile-list` for giving file of input files list (one file per line), if given, they are appended to files from cli arguments
    - `csvtk join`:
        - reenable flag `-i/--ignore-case`. [#99](https://github.com/shenwei356/csvtk/issues/99)
        - **outer join is supported**. [#23](https://github.com/shenwei356/csvtk/issues/23)
        - new flag `-L/--left-join`: left join, equals to -k/--keep-unmatched, exclusive with `--outer-join`
        - new flag `-O/--outer-join`: outer join, exclusive with --left-join
        - rename flag `--fill` to `--na`.
    - `csvtk filter2`: fix bug when column names start with digits, e.g., `1000g2015aug`. Thank @VorontsovIE ([#44](https://github.com/shenwei356/csvtk/issues/44))
    - `csvtk concat`: allow one input file. [#98](https://github.com/shenwei356/csvtk/issues/98)
    - `csvtk mutate`: new flag `-R/--remove` for removing input column.
- [csvtk v0.19.1](https://github.com/shenwei356/csvtk/releases/tag/v0.19.1)
[![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.19.1/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.19.1)
    - `csvtk`:
        - fix checking file existence.
        - show friendly error message when giving empty field like `csvtk cut -f a, b`.
    - `csvtk summary`: fix err of q1 and q3. [#90](https://github.com/shenwei356/csvtk/issues/90)
    - `csvtk version`: making checking update optional.
- [csvtk v0.19.0](https://github.com/shenwei356/csvtk/releases/tag/v0.19.0)
[![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.19.0/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.19.0)
    - [new commands by @bsipos](https://github.com/shenwei356/csvtk/pull/84):
        - `watch`: online monitoring and histogram of selected field.
        - `corr`: calculate Pearson correlation between numeric columns.
        - `cat`: stream file and report progress.
    - `csvtk split`: fix bug of repeatedly output header line when number of output files exceed value of `--buf-groups`. [#83](https://github.com/shenwei356/csvtk/issues/83)
    - `csvtk plot hist`: new option `--percentiles` to add percentiles to histogram x label. [#88](https://github.com/shenwei356/csvtk/pull/88)
- [csvtk v0.18.2](https://github.com/shenwei356/csvtk/releases/tag/v0.18.2)
[![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.18.2/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.18.2)
    - `csvtk replace/rename2/splitxlsx`: fix flag conflicts with global flag `-I` since v0.18.0.
    - `csvtk replace/rename2`: removing shorthand flag `-I` for `--key-capt-idx`.
    - `csvtk splitxlsx`: changing shorthand flag of `--sheet-index` from `-I` to `-N`.
- [csvtk v0.18.1](https://github.com/shenwei356/csvtk/releases/tag/v0.18.1)
[![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.18.1/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.18.1)
    - `csvtk sort`: fix mutiple-key-sort containing natural order sorting. [#79](https://github.com/shenwei356/csvtk/issues/79)
    - `csvtk xlsx2csv`: reacts to global flags `-t`, `-T`, `-D` and `-E`. [#78](https://github.com/shenwei356/csvtk/issues/78)
- [csvtk v0.18.0](https://github.com/shenwei356/csvtk/releases/tag/v0.18.0)
[![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.18.0/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.18.0)
    - `csvtk`: add new flag `--ignore-illegal-row` to skip illegal rows. [#72](https://github.com/shenwei356/csvtk/issues/72)
    - `csvtk summary`: add more textual/numeric operations. [#64](https://github.com/shenwei356/csvtk/issues/64)
    - `csvtk sort`: fix bug for sorting by columns with empty values. [#70](https://github.com/shenwei356/csvtk/issues/70)
    - `csvtk grep`: add new flag `--delete-matched` to delete a pattern right after being matched, this keeps the firstly matched data and speedups when using regular expressions. [#77](https://github.com/shenwei356/csvtk/issues/77)
- [csvtk v0.17.0](https://github.com/shenwei356/csvtk/releases/tag/v0.17.0)
[![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.17.0/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.17.0)
    - new command: `csvtk add-header` and `csvtk del-header` for adding/deleting column names. [#62](https://github.com/shenwei356/csvtk/issues/62)
- [csvtk v0.16.0](https://github.com/shenwei356/csvtk/releases/tag/v0.16.0)
[![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.16.0/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.16.0)
    - new command: `csvtk csv2json`: convert CSV to JSON format.
    - remove comand: `csvtk stats2`.
    - new command `csvtk summary`: summary statistics of selected digital fields (groupby group fields), [usage and examples](https://bioinf.shenwei.me/csvtk/usage/#stats). [#59](https://github.com/shenwei356/csvtk/issues/59)
    - `csvtk replace`: add flag `--nr-width`: minimum width for {nr} in flag -r/--replacement. e.g., formating "1" to "001" by `--nr-width 3` (default 1)
    - `csvtk rename2/replace`: add flag `-A, --kv-file-all-left-columns-as-value`, for treating all columns except 1th one as value for kv-file with more than 2 columns. [#56](https://github.com/shenwei356/csvtk/issues/56)
- [csvtk v0.15.0](https://github.com/shenwei356/csvtk/releases/tag/v0.15.0)
[![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.15.0/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.15.0)
    - `csvtk`: add global flag `-E/--ignore-empty-row` to skip empty row. [#50](https://github.com/shenwei356/csvtk/issues/50)
    - `csvtk mutate2`: add flag `-s/--digits-as-string` for not converting big digits into scientific notation. [#46](https://github.com/shenwei356/csvtk/issues/46)
    - `csvtk sort`: add support for sorting in natural order. [#49](https://github.com/shenwei356/csvtk/issues/49)
- [csvtk v0.14.0](https://github.com/shenwei356/csvtk/releases/tag/v0.14.0)
    [![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.14.0/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.14.0)
    - `csvtk`: **supporting multi-line fields by replacing [multicorecsv](https://github.com/mzimmerman/multicorecsv ) with standard library [encoding/csv](https://golang.org/pkg/encoding/csv/),
    while losing [support for metaline](https://github.com/shenwei356/csvtk/issues/13) which was supported since v0.7.0**. It also gain a little speedup.
    - `csvtk sample`: add flag `-n/--line-number` to print line number as the first column ("n")
    - `csvtk filter2`: fix bug when column names start with digits, e.g., `1000g2015aug` ([#44](https://github.com/shenwei356/csvtk/issues/44))
    - `csvtk rename2`: add support for similar repalecement symbols `{kv} and {nr}` in `csvtk replace`
- [csvtk v0.13.0](https://github.com/shenwei356/csvtk/releases/tag/v0.13.0)
[![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.13.0/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.13.0)
    - new command `concat` for concatenating CSV/TSV files by rows [#38](https://github.com/shenwei356/csvtk/issues/38)
    - `csvtk`: add support for environment variables for frequently used global flags [#39](https://github.com/shenwei356/csvtk/issues/39)
        - `CSVTK_T` for flag `-t/--tabs`
        - `CSVTK_H` for flag `-H/--no-header-row`
    - `mutate2`: add support for eval expression WITHOUT column index symbol, so we can add some string constants [#37](https://github.com/shenwei356/csvtk/issues/37)
    - `pretty`: better support for files with duplicated column names
- [csvtk v0.12.0](https://github.com/shenwei356/csvtk/releases/tag/v0.12.0)
[![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.12.0/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.12.0)
    - new command `collapse`: collapsing one field with selected fields as keys
    - `freq`: keeping orignal order of keys by default
    - `split`:
        - performance improvement
        - add option `-G/--out-gzip` for forcing output gzipped file
- [csvtk v0.11.0](https://github.com/shenwei356/csvtk/releases/tag/v0.11.0)
[![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.11.0/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.11.0)
    - add command `split` to split CSV/TSV into multiple files according to column values
    - add command `splitxlxs` to split XLSX sheet into multiple sheets according to column values
    - `csvtk`, automatically check BOM (byte-order mark) and discard it
- [csvtk v0.10.0](https://github.com/shenwei356/csvtk/releases/tag/v0.10.0)
[![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.10.0/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.10.0)
    - add subcommand `xlsx2csv` to convert XLSX to CSV format
    - `grep`, `filter`, `filter2`: add flag `-n/--line-number` to print line-number as the first column
    - `cut`: add flag `-i/--ignore-case` to ignore case of column name
- [csvtk v0.9.1](https://github.com/shenwei356/csvtk/releases/tag/v0.9.1)
[![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.9.1/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.9.1)
    - `csvtk replace`: fix bug when replacing with key-value pairs brought in v0.8.0
- [csvtk v0.9.0](https://github.com/shenwei356/csvtk/releases/tag/v0.9.0)
[![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.9.0/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.9.0)
    - add subcommand `csvtk mutate2`: create new column from selected fields by **awk-like arithmetic/string expressions**
    - add new command `genautocomplete` to generate **shell autocompletion** script!
- [csvtk v0.8.0](https://github.com/shenwei356/csvtk/releases/tag/v0.8.0)
[![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.8.0/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.8.0)
    - **new command `csvtk gather` for gathering columns into key-value pairs**.
    - `csvtk sort`: support **sorting by user-defined order**.
    - fix bug of *unselecting field*: wrongly reporting error of fields not existing.
    affected commands: `cut`, `filter`, `fitler2`, `freq`, `grep`, `inter`, `mutate`,
    `rename`, `rename2`, `replace`, `stats2`, `uniq`.
    - update help message of flag `-F/--fuzzy-fields`.
    - update help message of global flag `-t`, which overrides both `-d` and `-D`.
      If you want other delimiter for tabular input, use `-t $'\t' -D "delimiter"`.
- [csvtk v0.7.1](https://github.com/shenwei356/csvtk/releases/tag/v0.7.1)
[![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.7.1/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.7.1)
    - `csvtk plot box` and `csvtk plot line`: fix bugs for special cases of input
    - compile with go1.8.1
- [csvtk v0.7.0](https://github.com/shenwei356/csvtk/releases/tag/v0.7.0)
[![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.7.0/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.7.0)
    - fig bug of "stricter field checking" in v0.6.0 and v0.6.1 when using flag `-F/--fuzzy-fields`
    - `csvtk pretty` and `csvtk csv2md`: add attention that
      these commands treat the first row as header line and require them to be unique.
    - `csvtk stat` renamed to `csvtk stats`, old name is still available as an alias.
    - `csvtk stat2` renamed to `csvtk stats2`, old name is still available as an alias.
    - [issues/13](https://github.com/shenwei356/csvtk/issues/13) **seamlessly support for data with meta line of separator declaration used by MS Excel**.
- [csvtk v0.6.1](https://github.com/shenwei356/csvtk/releases/tag/v0.6.1)
[![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.6.1/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.6.1)
    - `csvtk cut`: minor bug: panic when no fields given. i.e., `csvtk cut`.
All relevant commands have been fixed.
- [csvtk v0.6.0](https://github.com/shenwei356/csvtk/releases/tag/v0.6.0)
[![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.6.0/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.6.0)
    - `csvtk grep`: **large performance improvement by discarding goroutine** (multiple threads),
      and **keeping output in order of input**.
    - Better column name checking and **stricter field checking,
      fields out of range are not ignored now**.
      Affected commands include `cut`, `filter`, `freq`, `grep`, `inter`, `mutate`,
      `rename`, `rename2`, `replace`, `stat2`, and `uniq`.
    - **New command: `csvtk filter2`, filtering rows by arithmetic/string expressions like `awk`**.
- [csvtk v0.5.0](https://github.com/shenwei356/csvtk/releases/tag/v0.5.0)
[![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.5.0/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.5.0)
    - `csvtk cut`: delete flag `-n/--names`, move it to a new command `csvtk headers`
    - new command: `csvtk headers`
    - new command: `csvtk head`
    - new command: `csvtk sample`
- [csvtk v0.4.6](https://github.com/shenwei356/csvtk/releases/tag/v0.4.6)
[![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.4.6/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.4.6)
    - `csvtk grep`: fix result highlight when flag `-v` is on.
- [csvtk v0.4.5](https://github.com/shenwei356/csvtk/releases/tag/v0.4.5)
[![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.4.5/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.4.5)
    - `csvtk join`: support the 2nd or later files with entries with same ID.
- [csvtk v0.4.4](https://github.com/shenwei356/csvtk/releases/tag/v0.4.4)
[![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.4.4/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.4.4)
    - add command `csvtk freq`: frequencies of selected fields
    - add lots of examples in [usage page](http://bioinf.shenwei.me/csvtk/usage/)
- [csvtk v0.4.3](https://github.com/shenwei356/csvtk/releases/tag/v0.4.3)
[![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.4.3/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.4.3)
    - improvement of using experience: flag `-n` is not required anymore when flag `-H` in `csvtk mutate`
- [csvtk v0.4.2](https://github.com/shenwei356/csvtk/releases/tag/v0.4.2)
[![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.4.2/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.4.2)
    - fix highlight bug of `csvtk grep`: if the pattern matches multiple parts,
    the text will be wrongly edited.
    - changes: disable highlight when pattern file given.
    - change the default output of all ploting commands to STDOUT, now you can
    pipe the image to "display" command of Imagemagic.
- [csvtk v0.4.1](https://github.com/shenwei356/csvtk/releases/tag/v0.4.1)
[![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.4.1/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.4.1)
    - Nothing changed. Just fix the links due to inappropriate deployment of v0.4.0
- [csvtk v0.4.0](https://github.com/shenwei356/csvtk/releases/tag/v0.4.0)
[![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.4.0/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.4.0)
    - add flag for `csvtk replace`: `-K` (`--keep-key`) keep the key as value when
    no value found for the key. This is open in default in previous versions.
- [csvtk v0.3.9](https://github.com/shenwei356/csvtk/releases/tag/v0.3.9)
  [![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.3.9/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.4.0)
    - fix bug: header row incomplete in `csvtk sort` result
- [csvtk v0.3.8.1](https://github.com/shenwei356/csvtk/releases/tag/v0.3.8.1)
  [![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.3.8.1/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.3.8.1)
    - fix bug of flag parsing library [pflag](https://github.com/spf13/pflag),
    [detail](https://github.com/spf13/pflag/pull/98).
    The bug affected the `csvtk grep -r -p`, when value of `-p` contain "[" and "]"
    at the beginning or end, they are wrongly parsed.
- [csvtk v0.3.8](https://github.com/shenwei356/csvtk/releases/tag/v0.3.8)
  [![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.3.8/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.3.8)
    - new feature: `csvtk cut` supports ordered fields output. e.g., `csvtk cut -f 2,1`
      outputs the 2nd column in front of 1th column.
    - new commands: `csvtk plot` can plot three types of plots by subcommands:
        - `csvtk plot hist`: histogram
        - `csvtk plot box`: boxplot
        - `csvtk plot line`: line plot and scatter plot
- [csvtk v0.3.7](https://github.com/shenwei356/csvtk/releases/tag/v0.3.7)
  [![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.3.7/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.3.7)
    - fix a serious bug of using negative field of column name, e.g. `-f "-id"`
- [csvtk v0.3.6](https://github.com/shenwei356/csvtk/releases/tag/v0.3.6)
  [![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.3.6/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.3.6)
    - `csvtk replace` support replacement symbols `{nr}` (record number)
      and `{kv}` (corresponding value of the key ($1) by key-value file)
- [csvtk v0.3.5.2](https://github.com/shenwei356/csvtk/releases/tag/v0.5.2)
  [![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.3.5.2/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.3.5.2)
    - add flag `--fill` for `csvtk join`, so we can fill the unmatched data
    - fix typo
- [csvtk v0.3.5.1](https://github.com/shenwei356/csvtk/releases/tag/v0.3.5.1)
  [![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.3.5.1/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.3.5.1)
    - fix minor bug of reading lines ending with `\r\n` from a dependency package
- [csvtk v0.3.5](https://github.com/shenwei356/csvtk/releases/tag/v0.3.5)
  [![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.3.5/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.3.5)
    - fix minor bug of `csv2md`
    - add subcommand `version` which could check for update
- [csvtk v0.3.4](https://github.com/shenwei356/csvtk/releases/tag/v0.3.4)
  [![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.3.4/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.3.4)
    - fix bug of `csvtk replace` that head row should not be edited.
- [csvtk v0.3.3](https://github.com/shenwei356/csvtk/releases/tag/v0.3.3)
  [![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.3.3/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.3.3)
    - fix bug of `csvtk grep -t -P`
- [csvtk v0.3.2](https://github.com/shenwei356/csvtk/releases/tag/v0.3.2)
  [![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.3.2/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.3.2)
    - fix bug of `inter`
- [csvtk v0.3.1](https://github.com/shenwei356/csvtk/releases/tag/v0.3.1)
  [![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.3.1/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.3.1)
    - add support of search multiple fields for `grep`
- [csvtk v0.3](https://github.com/shenwei356/csvtk/releases/tag/v0.3)
  [![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.3/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.3)
    - add subcommand `csv2md`
- [csvtk v0.2.9](https://github.com/shenwei356/csvtk/releases/tag/v0.2.9)
  [![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.2.9/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.2.9)
    - add more flags to subcommand `pretty`
    - fix bug of `csvtk cut -n`
    - add subcommand `filter`
- [csvtk v0.2.8](https://github.com/shenwei356/csvtk/releases/tag/v0.2.8)
  [![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.2.8/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.2.8)
    - add subcommand `pretty` -- convert CSV to readable aligned table
- [csvtk v0.2.7](https://github.com/shenwei356/csvtk/releases/tag/v0.2.7)
  [![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.2.7/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.2.7)
    - fix highlight failing in windows
- [csvtk v0.2.6](https://github.com/shenwei356/csvtk/releases/tag/v0.2.6)
  [![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.2.6/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.2.6)
    - fix one error message of `grep`
    - highlight matched fields in result of `grep`
- [csvtk v0.2.5](https://github.com/shenwei356/csvtk/releases/tag/v0.2.5)
  [![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.2.5/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.2.5)
    - fix bug of `stat` that failed to considerate files with header row
    - add subcommand `stat2` - summary of selected number fields
    - make the output of `stat` prettier
- [csvtk v0.2.4](https://github.com/shenwei356/csvtk/releases/tag/v0.2.4)
  [![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.2.4/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.2.4)
    - fix bug of handling comment lines
    - add some notes before using csvtk
- [csvtk v0.2.3](https://github.com/shenwei356/csvtk/releases/tag/v0.2.3)
  [![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.2.3/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.2.3)
    - add flag `--colnames` to `cut`
    - flag `-f` (`--fields`) of `join` supports single value now
- [csvtk v0.2.2](https://github.com/shenwei356/csvtk/releases/tag/v0.2.2)
  [![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.2.2/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.2.2)
    - add flag `--keep-unmathed` to `join`
- [csvtk v0.2](https://github.com/shenwei356/csvtk/releases/tag/v0.2)
  [![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.2/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.2)
    - finish almost functions
- [csvtk v0.2.1](https://github.com/shenwei356/csvtk/releases/tag/v0.2.1)
  [![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.2.1/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.2.1)
    - fix bug of `mutate`
