
What
----
http://wunderdog.fi/koodaus-pahkina-kesa-2015

How
---

Build (downloads resources)

```
make build
```

Run

```
$ time ./muhkeat 
                      Input file: alastalon_salissa.txt
              Characters handled: abcdefghijklmnopqrstuvwzyxåäö
 Unique (case insensitive) words: 57213
            Unique sets of chars: 22992

 Top pairs found (weight 21)
-----------------------------
kirjoituspöydille vahingonkamfiili
kirjoituspöydän vahingonpillastumisen
kirjoituspöydän vahingonkamfiili
kirjoituspöydältä vahingonkamfiili
kirjoituspöydältä vahingonpillastumisen
köydenpingottuvaa järjensyhyttelemisen
vahingonpillastumisen keräjäpöydän
vahingonpillastumisen kirjoituspöydästä
vahingonkamfiili kirjoituspöydästä

real    0m0.400s
user    0m0.396s
sys     0m0.004s
```

Why
---
I wanted to test Go with something simpleish.

Who
---
https://github.com/zemm
