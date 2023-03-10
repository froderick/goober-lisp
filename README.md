# goober-lisp

This is a program I wrote during a hackathon with the goal of learning Golang
by writing a simple lisp interpreter.

Please don't write anything useful in this, it will probably take your hand
off. But I highly recommend Golang, and Lisp!

## getting started

You can run the following from the root of the repo to get the repl running.

```
$ ./repl.sh
```

To get a sense of what the simple language is capable of, take a look at the [core.el](core.el) where the builtins of the language are defined.

Here's the fibonacci sequence:

```lisp
user> (defn fib (n)
  (let (fib-n (fn (x n)
                (if (< (count x) n) 
                  (recur (cons (+ (first x) (second x)) x) n)
                  (reverse x))))
    (fib-n '(1 0) n)))

user> (fib 10)
(0 1 1 2 3 5 8 13 21 34)
```
