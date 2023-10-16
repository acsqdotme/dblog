# dblog: 🐹 + 📜🪶 = 🌐✍️

Go pkg for database system for some of the sites I'm building. Originally part
of [angel-castaneda.com git repo](https://git.acsq.me/angel-castaneda.com).

## how it works

dblog works by scanning a root posts directory that has the following structure:

```console
user@host:path/to/posts $ ls
cool-post/
epic-post/
awesome-post/
... # more post directories corresponding to url/filename of that post
```

then you can [ls] into any one of them and find:

```console
user@host:path/to/posts/cool-post $ ls
meta.yml # only one that matters
photos/ # scanner ignores anything else in dir
post.html
post.md
```

and [meta.yml] just needs to carry some metadata about the blog post to be
useful (example of directory structure [here](./example-blog-post))

Soon enough, I'll it scan for tags too and be a bit smarter about updating and
deleting posts not found in the directory (also I guess I'll have to make the
directory and database locales changeable.)

## converting to a git submodule

This was a way for me to learn git submodules and put some progress here.

I followed this simple [stack overflow
guide](https://stackoverflow.com/a/73598455/21316874) to move my git history
over.

Then I went to my main repo, removed the original directory, and ran:

```console
$ git submodule add https://git.sr.ht/~acsqdotme/dblog
```

That cloned in this new directory and a `.gitmodules` file:

```git
[submodule "dblog"]
	path = dblog
	url = https://git.sr.ht/~acsqdotme/dblog
```

As always, more details can be found from [git
themselves](https://git-scm.com/book/en/v2/Git-Tools-Submodules).

## License

This project is licensed under the LGPLv3. Check [`LICENSE`](./LICENSE) for
details.
