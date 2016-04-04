# csvtk

Another cross-platform, efficient and practical CSV/TSV tool kit.

## Features

- **Cross-platform** (Linux/Windows/Mac OS X/OpenBSD/FreeBSD)
- **Light weight and out-of-the-box, no dependencies, no compilation, no configuration**
- **Fast**,  **multiple-CPUs supported**
- **Practical functions supported by N subcommands**
- **Support STDIN and gziped input/output file, easy being used in pipe**

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
-  `split` split data to multiple files by values of selected fields
-  `grep` grep data by selected fields with patterns
-  `filter` filter data by values of selected fields, supporting math/string expression
-  `join` join multiple CSV files by selected fields
-  `uniq` unique data without sorting
-  `inter` intersection of multiple files

**Edit**

-  `replace` replace data of selected fields by regular expression
-  `mutate` create new columns from selected fields by regular expression

**Ordering**

-  `sort` sort by selected fields

## Compared to `csvkit`

[csvkit](http://csvkit.readthedocs.org/en/540/)

TODO

## Contact

Email me for any problem when using `csvtk`. shenwei356(at)gmail.com

[Create an issue](https://github.com/shenwei356/csvtk/issues) to report bugs,
propose new functions or ask for help.

## License

[MIT License](https://github.com/shenwei356/csvtk/blob/master/LICENSE)
