## cwgo completion zsh

Generate the autocompletion script for zsh

### Synopsis

Generate the autocompletion script for the zsh shell.

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

	echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions in your current shell session:

	source <(cwgo completion zsh); compdef _cwgo cwgo

To load completions for every new session, execute once:

#### Linux:

	cwgo completion zsh > "${fpath[1]}/_cwgo"

#### macOS:

	cwgo completion zsh > $(brew --prefix)/share/zsh/site-functions/_cwgo

You will need to start a new shell for this setup to take effect.


```
cwgo completion zsh [flags]
```

### Options

```
  -h, --help              help for zsh
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.cwgo.yaml)
  -d, --debug           Print output instead of creating files
```

### SEE ALSO

* [cwgo completion](cwgo_completion.md)	 - Generate the autocompletion script for the specified shell

###### Auto generated by spf13/cobra on 9-Jun-2023