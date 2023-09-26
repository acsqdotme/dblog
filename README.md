# dblog

Go pkg for database system for some of the sites I'm building. Originally part
of [angel-castaneda.com git repo](https://git.acsq.me/angel-castaneda.com).

## 🐹 + 📜🪶 = ✍️

I'm about to update a whole ton with this package to have auto post directory
scanning. Will put details on how it works once it's done.

## converting to a git submodule

This is a way for me to learn git submodules and put some progress here.

I followed this simple
[stackoverflow guide](https://stackoverflow.com/a/73598455/21316874) to move my
git history over.

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
