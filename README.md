OpenFaaS Cloud beta
====================

OpenFaaS Cloud uses functions to perform CI/CD for functions hosted on GitHub.


Build script:

```
faas-cli build -f stack.yml --parallel=4 && faas-cli push -f stack.yml --parallel=4 && faas-cli deploy -f stack.yml
```

