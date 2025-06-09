###

#### bug
```shell
# this works in standard flag, 
bin/template < /tmp/sample.yml --set podLabels='{"hello":"world"}'

```
but not in spf13/pflag, it throws error
```shell
bin/template < /tmp/sample.yml --set podLabels='{"hello":"world"}'

invalid argument "podLabels={\"hello\":\"world\"}" for "-s, --set" flag: parse error on line 1, column 12: bare " in non-quoted-field
Usage of bin/template:
  -s, --set strings       --set key=value --set key2=value2
  -t, --template string   -t <template-file>
invalid argument "podLabels={\"hello\":\"world\"}" for "-s, --set" flag: parse error on line 1, column 12: bare " in non-quoted-field
```
