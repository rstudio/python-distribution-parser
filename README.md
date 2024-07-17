# Python Distribution Parser

Have you ever wanted to parse Python distributions like `.whl` files from Go before? Wait, you haven't? Huh.

## Purpose

The main goal is to provide the ability to parse the metadata of a Python distribution. This almost entirely matches what `twine` does, feel free to compare the code here:

https://github.com/pypa/twine/blob/afc37f8b26ed06ccd104f6724f293f657b9b7f15/twine/commands/upload.py#L93

This implementation supports parsing only `.tar.gz` and `.whl` files (also know as `sdist` and `bdist_whl`). Currently, `.exe` (`bdist_wininst`), and `.egg` (`bdist_egg`) are not supported.

## Installation

```sh
$ go get github.com/rstudio/python-distribution-parser
```

## Usage

The main entrypoint is the `Parse(path string) ([]*packages.PackageFile, error)` function. This takes in a file or directory path and will return an array of `*PackageFile` objects based on the distribution(s) passed in. Here it is in action:

```go
package main
import (
    "encoding/json"
    "fmt"

    pdp "github.com/rstudio/python-distribution-parser"
)


func main() {
    filePath := "/path/to/appdirs/dist/appdirs-1.4.4-py2.py3-none-any.whl"
    packages, err := pdp.Parse(filePath)
    if err != nil {
        log.Fatalf("Error parsing directory %s: %v", filePath, err)
    }

    jsonData, err := json.MarshalIndent(packages, "", "  ")
    if err != nil {
        log.Fatalf("Error marshalling to JSON: %v", err)
    }

    fmt.Println(string(jsonData))
}
```

And here is example output:

```json
[
  {
    "filename": "/path/to/appdirs/dist/appdirs-1.4.4-py2.py3-none-any.whl",
    "base_filename": "appdirs-1.4.4-py2.py3-none-any.whl",
    "metadata": {
      "metadata_version": "2.1",
      "name": "appdirs",
      "version": "1.4.4",
      "platforms": null,
      "supported_platforms": null,
      "summary": "A small Python module for determining appropriate platform-specific dirs, e.g. a \"user data dir\".",
      "description": "\n.. image:: https://secure.travis-ci.org/ActiveState/appdirs.png\n    :target: https://travis-ci.org/ActiveState/appdirs\n\nNote: This project has been officially deprecated. You may want to check out https://pypi.org/project/platformdirs/ which is a more active fork of appdirs. Thanks to everyone who has used appdirs. Shout out to ActiveState for the time they gave their employees to work on this over the years.\n\nthe problem\n===========\n\nWhat directory should your app use for storing user data? If running on macOS, you\nshould use::\n\n    ~/Library/Application Support/\u003cAppName\u003e\n\nIf on Windows (at least English Win XP) that should be::\n\n    C:\\Documents and Settings\\\u003cUser\u003e\\Application Data\\Local Settings\\\u003cAppAuthor\u003e\\\u003cAppName\u003e\n\nor possibly::\n\n    C:\\Documents and Settings\\\u003cUser\u003e\\Application Data\\\u003cAppAuthor\u003e\\\u003cAppName\u003e\n\nfor `roaming profiles \u003chttps://docs.microsoft.com/en-us/previous-versions/windows/it-pro/windows-vista/cc766489(v=ws.10)\u003e`_ but that is another story.\n\nOn Linux (and other Unices) the dir, according to the `XDG\nspec \u003chttps://standards.freedesktop.org/basedir-spec/basedir-spec-latest.html\u003e`_, is::\n\n    ~/.local/share/\u003cAppName\u003e\n\n\n``appdirs`` to the rescue\n=========================\n\nThis kind of thing is what the ``appdirs`` module is for. ``appdirs`` will\nhelp you choose an appropriate:\n\n- user data dir (``user_data_dir``)\n- user config dir (``user_config_dir``)\n- user cache dir (``user_cache_dir``)\n- site data dir (``site_data_dir``)\n- site config dir (``site_config_dir``)\n- user log dir (``user_log_dir``)\n\nand also:\n\n- is a single module so other Python packages can include their own private copy\n- is slightly opinionated on the directory names used. Look for \"OPINION\" in\n  documentation and code for when an opinion is being applied.\n\n\nsome example output\n===================\n\nOn macOS::\n\n    \u003e\u003e\u003e from appdirs import *\n    \u003e\u003e\u003e appname = \"SuperApp\"\n    \u003e\u003e\u003e appauthor = \"Acme\"\n    \u003e\u003e\u003e user_data_dir(appname, appauthor)\n    '/Users/trentm/Library/Application Support/SuperApp'\n    \u003e\u003e\u003e site_data_dir(appname, appauthor)\n    '/Library/Application Support/SuperApp'\n    \u003e\u003e\u003e user_cache_dir(appname, appauthor)\n    '/Users/trentm/Library/Caches/SuperApp'\n    \u003e\u003e\u003e user_log_dir(appname, appauthor)\n    '/Users/trentm/Library/Logs/SuperApp'\n\nOn Windows 7::\n\n    \u003e\u003e\u003e from appdirs import *\n    \u003e\u003e\u003e appname = \"SuperApp\"\n    \u003e\u003e\u003e appauthor = \"Acme\"\n    \u003e\u003e\u003e user_data_dir(appname, appauthor)\n    'C:\\\\Users\\\\trentm\\\\AppData\\\\Local\\\\Acme\\\\SuperApp'\n    \u003e\u003e\u003e user_data_dir(appname, appauthor, roaming=True)\n    'C:\\\\Users\\\\trentm\\\\AppData\\\\Roaming\\\\Acme\\\\SuperApp'\n    \u003e\u003e\u003e user_cache_dir(appname, appauthor)\n    'C:\\\\Users\\\\trentm\\\\AppData\\\\Local\\\\Acme\\\\SuperApp\\\\Cache'\n    \u003e\u003e\u003e user_log_dir(appname, appauthor)\n    'C:\\\\Users\\\\trentm\\\\AppData\\\\Local\\\\Acme\\\\SuperApp\\\\Logs'\n\nOn Linux::\n\n    \u003e\u003e\u003e from appdirs import *\n    \u003e\u003e\u003e appname = \"SuperApp\"\n    \u003e\u003e\u003e appauthor = \"Acme\"\n    \u003e\u003e\u003e user_data_dir(appname, appauthor)\n    '/home/trentm/.local/share/SuperApp\n    \u003e\u003e\u003e site_data_dir(appname, appauthor)\n    '/usr/local/share/SuperApp'\n    \u003e\u003e\u003e site_data_dir(appname, appauthor, multipath=True)\n    '/usr/local/share/SuperApp:/usr/share/SuperApp'\n    \u003e\u003e\u003e user_cache_dir(appname, appauthor)\n    '/home/trentm/.cache/SuperApp'\n    \u003e\u003e\u003e user_log_dir(appname, appauthor)\n    '/home/trentm/.cache/SuperApp/log'\n    \u003e\u003e\u003e user_config_dir(appname)\n    '/home/trentm/.config/SuperApp'\n    \u003e\u003e\u003e site_config_dir(appname)\n    '/etc/xdg/SuperApp'\n    \u003e\u003e\u003e os.environ['XDG_CONFIG_DIRS'] = '/etc:/usr/local/etc'\n    \u003e\u003e\u003e site_config_dir(appname, multipath=True)\n    '/etc/SuperApp:/usr/local/etc/SuperApp'\n\n\n``AppDirs`` for convenience\n===========================\n\n::\n\n    \u003e\u003e\u003e from appdirs import AppDirs\n    \u003e\u003e\u003e dirs = AppDirs(\"SuperApp\", \"Acme\")\n    \u003e\u003e\u003e dirs.user_data_dir\n    '/Users/trentm/Library/Application Support/SuperApp'\n    \u003e\u003e\u003e dirs.site_data_dir\n    '/Library/Application Support/SuperApp'\n    \u003e\u003e\u003e dirs.user_cache_dir\n    '/Users/trentm/Library/Caches/SuperApp'\n    \u003e\u003e\u003e dirs.user_log_dir\n    '/Users/trentm/Library/Logs/SuperApp'\n\n\n    \nPer-version isolation\n=====================\n\nIf you have multiple versions of your app in use that you want to be\nable to run side-by-side, then you may want version-isolation for these\ndirs::\n\n    \u003e\u003e\u003e from appdirs import AppDirs\n    \u003e\u003e\u003e dirs = AppDirs(\"SuperApp\", \"Acme\", version=\"1.0\")\n    \u003e\u003e\u003e dirs.user_data_dir\n    '/Users/trentm/Library/Application Support/SuperApp/1.0'\n    \u003e\u003e\u003e dirs.site_data_dir\n    '/Library/Application Support/SuperApp/1.0'\n    \u003e\u003e\u003e dirs.user_cache_dir\n    '/Users/trentm/Library/Caches/SuperApp/1.0'\n    \u003e\u003e\u003e dirs.user_log_dir\n    '/Users/trentm/Library/Logs/SuperApp/1.0'\n\n\n\nappdirs Changelog\n=================\n\nUNRELEASED\n----------\n- Add Python 3.7 support\n- Remove support for end-of-life Pythons 2.6, 3.2, and 3.3\n\nappdirs 1.4.4\n-------------\n- [PR #92] Don't import appdirs from setup.py which resolves issue #91\n- [PR #100] Corrects the config directory on OSX/macOS, which resolves issue #63.\n\nProject officially classified as Stable which is important\nfor inclusion in other distros such as ActivePython.\n\nappdirs 1.4.3\n-------------\n- [PR #76] Python 3.6 invalid escape sequence deprecation fixes\n- Fix for Python 3.6 support\n\nappdirs 1.4.2\n-------------\n- [PR #84] Allow installing without setuptools\n- [PR #86] Fix string delimiters in setup.py description\n- Add Python 3.6 support\n\nappdirs 1.4.1\n-------------\n- [issue #38] Fix _winreg import on Windows Py3\n- [issue #55] Make appname optional\n\nappdirs 1.4.0\n-------------\n- [PR #42] AppAuthor is now optional on Windows\n- [issue 41] Support Jython on Windows, Mac, and Unix-like platforms. Windows\n  support requires `JNA \u003chttps://github.com/twall/jna\u003e`_.\n- [PR #44] Fix incorrect behaviour of the site_config_dir method\n\nappdirs 1.3.0\n-------------\n- [Unix, issue 16] Conform to XDG standard, instead of breaking it for\n  everybody\n- [Unix] Removes gratuitous case mangling of the case, since \\*nix-es are\n  usually case sensitive, so mangling is not wise\n- [Unix] Fixes the utterly wrong behaviour in ``site_data_dir``, return result\n  based on XDG_DATA_DIRS and make room for respecting the standard which\n  specifies XDG_DATA_DIRS is a multiple-value variable\n- [Issue 6] Add ``*_config_dir`` which are distinct on nix-es, according to\n  XDG specs; on Windows and Mac return the corresponding ``*_data_dir``\n\nappdirs 1.2.0\n-------------\n\n- [Unix] Put ``user_log_dir`` under the *cache* dir on Unix. Seems to be more\n  typical.\n- [issue 9] Make ``unicode`` work on py3k.\n\nappdirs 1.1.0\n-------------\n\n- [issue 4] Add ``AppDirs.user_log_dir``.\n- [Unix, issue 2, issue 7] appdirs now conforms to `XDG base directory spec\n  \u003chttps://standards.freedesktop.org/basedir-spec/basedir-spec-latest.html\u003e`_.\n- [Mac, issue 5] Fix ``site_data_dir()`` on Mac.\n- [Mac] Drop use of 'Carbon' module in favour of hardcoded paths; supports\n  Python3 now.\n- [Windows] Append \"Cache\" to ``user_cache_dir`` on Windows by default. Use\n  ``opinion=False`` option to disable this.\n- Add ``appdirs.AppDirs`` convenience class. Usage:\n\n        \u003e\u003e\u003e dirs = AppDirs(\"SuperApp\", \"Acme\", version=\"1.0\")\n        \u003e\u003e\u003e dirs.user_data_dir\n        '/Users/trentm/Library/Application Support/SuperApp/1.0'\n\n- [Windows] Cherry-pick Komodo's change to downgrade paths to the Windows short\n  paths if there are high bit chars.\n- [Linux] Change default ``user_cache_dir()`` on Linux to be singular, e.g.\n  \"~/.superapp/cache\".\n- [Windows] Add ``roaming`` option to ``user_data_dir()`` (for use on Windows only)\n  and change the default ``user_data_dir`` behaviour to use a *non*-roaming\n  profile dir (``CSIDL_LOCAL_APPDATA`` instead of ``CSIDL_APPDATA``). Why? Because\n  a large roaming profile can cause login speed issues. The \"only syncs on\n  logout\" behaviour can cause surprises in appdata info.\n\n\nappdirs 1.0.1 (never released)\n------------------------------\n\nStarted this changelog 27 July 2010. Before that this module originated in the\n`Komodo \u003chttps://www.activestate.com/komodo-ide\u003e`_ product as ``applib.py`` and then\nas `applib/location.py\n\u003chttps://github.com/ActiveState/applib/blob/master/applib/location.py\u003e`_ (used by\n`PyPM \u003chttps://code.activestate.com/pypm/\u003e`_ in `ActivePython\n\u003chttps://www.activestate.com/activepython\u003e`_). This is basically a fork of\napplib.py 1.0.1 and applib/location.py 1.0.1.\n",
      "keywords": "application directory log cache user",
      "home_page": "https://github.com/ActiveState/appdirs",
      "download_url": "",
      "author": "Trent Mick",
      "author_email": "trentm@gmail.com",
      "license": "MIT",
      "classifiers": [
        "Development Status :: 5 - Production/Stable",
        "Intended Audience :: Developers",
        "License :: OSI Approved :: MIT License",
        "Operating System :: OS Independent",
        "Programming Language :: Python",
        "Programming Language :: Python :: 2",
        "Programming Language :: Python :: 2.7",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.5",
        "Programming Language :: Python :: 3.6",
        "Programming Language :: Python :: 3.7",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: Implementation :: PyPy",
        "Programming Language :: Python :: Implementation :: CPython",
        "Topic :: Software Development :: Libraries :: Python Modules"
      ],
      "requires": null,
      "provides": null,
      "obsoletes": null,
      "maintainer": "Jeff Rouse",
      "maintainer_email": "jr@its.to",
      "requires_python": "\u003e=2.7, !=3.0.*, !=3.1.*, !=3.2.*, !=3.3.*, !=3.4.*",
      "requires_external": null,
      "requires_dist": null,
      "provides_dist": null,
      "obsoletes_dist": null,
      "project_urls": null,
      "provides_extra": null,
      "description_content_type": "",
      "dynamic": null,
      "filename": "/path/to/appdirs/dist/appdirs-1.4.4-py2.py3-none-any.whl",
      "base_filename": "appdirs-1.4.4-py2.py3-none-any.whl"
    },
    "python_version": "py2.py3",
    "file_type": "bdist_wheel",
    "safe_name": "appdirs",
    "signed_filename": "/path/to/appdirs/dist/appdirs-1.4.4-py2.py3-none-any.whl.asc",
    "signed_base_filename": "appdirs-1.4.4-py2.py3-none-any.whl.asc",
    "gpg_signature": null,
    "md5_digest": "3960d7797a1fb32d1229c59db8702c11",
    "sha2_digest": "2cf70fb91a350c7987e3410693978417056e5479aa476c0f4c68b3697d431776",
    "blake2_256_digest": "2a8c9ca4a072d43e00503f16d08db68d690644949ac3e4704ba6c2e7028e8402"
  }
]
```

## Linting

```bash
golangci-lint run --fix ./...
```

## Testing

```bash
go test ./...
```

To update snapshots with [cupaloy](https://github.com/bradleyjkemp/cupaloy):

```bash
UPDATE_SNAPSHOTS=true go test ./...
```
