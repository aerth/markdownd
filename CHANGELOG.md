# changelog

## markdownd 0.0.12
  * generate index file with '-index=gen'
  * minor improvements
  * refactor build scripts
  * use multi-stage docker build (*way* smaller image size)
  * add docker build/run examples

## markdownd 0.0.11
  * use github-flavored-markdown

## markdownd 0.0.10

  * Resolves filepath issue on windows systems (thank you @weaming)
  * Linux and OS X tests are being ran (see travis link on README.md)
  * More accurate response timer
  * Now making sure main ("/") -index file exists, warning if not
