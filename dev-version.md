0.7.2-dev

- fix bug of *unselecting field*: wrongly reporting error of fields not existing.
affected commands: `cut`, `filter`, `fitler2`, `freq`, `grep`, `inter`, `mutate`,
`rename`, `rename2`, `replace`, `stats2`, `uniq`.
- new command `csvtk gather` for gathering columns into key-value pairs.
- `csvtk sort`: support sort by user-defined order.
- update help message of flag `-F/--fuzzy-fields`.
- update help message of global flag `-t`, which overrides both `-d` and `-D`.
  If you want other delimiter for tabular input, use `-t $'\t' -D "delimiter"`.
