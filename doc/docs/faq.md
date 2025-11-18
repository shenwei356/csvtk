# Frequently Asked Questions

## The specification of CSV format

The CSV parser used by csvtk follows the [RFC4180](https://rfc-editor.org/rfc/rfc4180.html) specification.

## bare " in non-quoted-field

```
   5.  Each field may or may not be enclosed in double quotes (however
       some programs, such as Microsoft Excel, do not use double quotes
       at all).  If fields are not enclosed with double quotes, then
       double quotes may not appear inside the fields.  For example:

       "aaa","bbb","ccc" CRLF
       zzz,yyy,xxx

   6.  Fields containing line breaks (CRLF), double quotes, and commas
       should be enclosed in double-quotes.  For example:

       "aaa","b CRLF
       bb","ccc" CRLF
       zzz,yyy,xxx

   7.  If double-quotes are used to enclose fields, then a double-quote
       appearing inside a field must be escaped by preceding it with
       another double quote.  For example:

       "aaa","b""bb","ccc"
```

If a single double-quote exists in one non-quoted-field, an error will be reported. e.g,

    $ echo 'a,abc" xyz,d'
    a,abc" xyz,d

    $ echo 'a,abc" xyz,d' | csvtk cut -f 1-
    [ERRO] parse error on line 1, column 6: bare " in non-quoted-field

You can add the flag `-l/--lazy-quotes` to fix this.

    $ echo 'a,abc" xyz,d' | csvtk cut -f 1- -l
    a,"abc"" xyz",d

## extraneous or missing " in quoted-field

But for the situation below, `-l/--lazy-quotes` won't help:

    $ echo 'a,"abc" xyz,d'
    a,"abc" xyz,d

    $ echo 'a,"abc" xyz,d' | csvtk cut -f 1-
    [ERRO] parse error on line 1, column 7: extraneous or missing " in quoted-field

    $ echo 'a,"abc" xyz,d' | csvtk cut -f 1- -l
    a,"abc"" xyz,d
    "

    $ echo 'a,"abc" xyz,d' | csvtk cut -f 1- -l | csvtk dim
    file  num_cols  num_rows
    -            2         0

**You need to use [csvtk fix-quotes](https://bioinf.shenwei.me/csvtk/usage/#fix-quotes) (available in v0.29.0 or later versions)**:

    $ echo 'a,"abc" xyz,d' | csvtk fix-quotes
    a,"""abc"" xyz",d

    $ echo 'a,"abc" xyz,d' | csvtk fix-quotes | csvtk cut -f 1-
    a,"""abc"" xyz",d

    $ echo 'a,"abc" xyz,d' | csvtk fix-quotes | csvtk cut -f 1- | csvtk dim
    file  num_cols  num_rows
    -            3         0

Use [del-quotes](https://bioinf.shenwei.me/csvtk/usage/#del-quotes) if you need the original format after some operations.

    $ echo 'a,"abc" xyz,d' | csvtk fix-quotes | csvtk cut -f 1- | csvtk del-quotes
    a,"abc" xyz,d

## Environment variables

Environment variables for frequently used global flags:

  - `CSVTK_T` for flag `-t/--tabs`
  - `CSVTK_H` for flag `-H/--no-header-row`
  - `CSVTK_QUIET` for flag `--quiet`

You can also create a soft link named `tsvtk` for `csvtk`, which sets `-t/--tabs` by default.
