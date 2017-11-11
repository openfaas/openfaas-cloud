Build script:

```
faas-cli build -f stack.yml --parallel=4 && faas-cli push -f stack.yml --parallel=4 && faas-cli deploy -f stack.yml
```
