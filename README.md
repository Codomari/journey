# Journey
A blog engine written in Go, compatible with Ghost themes.

![Editor](https://raw.githubusercontent.com/kabukky/journey/gh-pages/images/journey.png)

## About
It's a fork of [https://github.com/kabukky/journey](https://github.com/kabukky/journey) blog engine which abandoned by author, but cloned to continue support for PR-s from enthusiasts who enjoy using it.

#### Easy to work with
Create or update your posts from any place and any device. Simply point your browser to yourblog.url/admin/, log in, and start typing away!

#### Lightweight and fast
Journey is still in an early stage of development. However, initial tests indicate that it is much faster at generating pages than Ghost running on Node.js. It also eats very little of your precious memory. For example: Testing it on Mac OS X, it takes about 3.5 MB of it and then happily carries on doing its job.

This slimness makes Journey an ideal candidate for setting up micro blogs or hosting it on low-end vps machines or micro computers such as the Raspberry Pi.

#### Deployable anywhere
[Download the release package](https://www.github.com/Codomari/journey/releases) for Linux (AMD64, i386, ARM), Mac OS X (AMD64, i386) or Windows (AMD64, i386) and start using Journey right away. Build Journey from source to make it work on a multitude of other operating systems!

## Questions?
Please contact me [by email](mailto:anar.k.jafarov@gmail.com)

## Troubleshooting

Please create a [New Issue](https://github.com/Codomari/journey/issues).

## Building from source
It's build as every golang app: 
```go
go mod download
go build -o journey
chmod +x journey
```

## Contributing to Journey
Pull requests are very much welcome. But please create them on the development branch. The master branch will only be updated for a new release.
