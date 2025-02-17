## Install

### Pre-requisites

 * appdmg (`npm install -g appdmg`)
 * xcpretty (`gem install xcpretty`)
 * Xcode command line tools

### Build

```sh
sh build.sh
```

## Overview

When the Keybase.app runs it checks for the following components and compares the bundled version with the installed and running versions to make sure it's installed and up to date:

- Service (Launch Agent)
- Privileged Helper Tool
- KBFS (Launch Agent)
- KBFuse (our custom osxfuse build, see [Fuse/kbfuse](https://github.com/keybase/client/tree/master/osx/Install/Fuse/kbfuse) for more details).
