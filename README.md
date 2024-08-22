# CodePlumber

## About The Project

**CodePlumber** is a CLI tool to monitor and interact with AWS CodePipelines.
If you use AWS CodePipeline, you are probably already aware that managing a large number of jobs can become quite difficult as the filtering capabilities are quite limited.

**Codeplumber** is designed to help you monitor the important jobs that matters to you and perform basic operations with ease.
Profiles can be created to target subsets of CodePipelines jobs, allowing you to quickly run and monitor them.

A fair warning, this project was primaraly motivated by the desire to learn and experiment with the [Charm](https://github.com/charmbracelet/) libraries. It works well for the jobs I handle and care about, but there are likely many bugs still lurking, and some use cases and configurations might not be well supported yet. Contributions are more than welcome!

## Getting Started

### Installation

TODO

### Configuration

The intended way to use this tool is to set a configuration file with default location being `$HOME/.config/codeplumber/config.yaml` in which all the different profiles are defined.

You can define as many profiles as you need and call them when required.
You can filter CodePipelines jobs based on their name and apply more granular filtering using AWS tags.

Currently, tag-only filters are not allowed because AWS does not provide an API for this. As a result, this operation is quite expensive, and API rate limits make it really slow. Therefore, name filtering is required to limit the number of CodePipelines to filter by tags.
This limitation also exists in the AWS console (no tag filtering is permitted there), so hopefully, you already have a good naming convention in place.
When using name and tag filters in a profile, the resulting operation will filter AWS CodePipelines that match all conditions.

```yaml
---
profiles:
  myProdDeployment:
    aws:
      region: us-east-1    # The AWS region to target (`us-east-1` by default)
      profile: my-profile  # The AWS profile to use (`default` by default)
    kind: codepipeline     # Only CodePipeline supported at the moment
    filters:
      name: myTeam-       # Filter CodePipelines that contain this string
      tags:
        environment: prod  # And this tag
        kind: deployment   # And this tag
```

This profile can then be called at any time with:

```bash
codeplumber load myProdDeployment
```

Alternatively, the same request can be crafted from the command line:

```bash
codeplumber run -p my-profile -r us-east-1 --filter-tags 'environment=prod,kind=deployment' myTeam-
```

Finally, when using profiles, another `name` filter can be added:

```bash
codeplumber load myProdDeployment myApp
```

This will filter all CodePipelines jobs matching all conditions of the `myProdDeployment` profile AND that contain also the string `myApp` as part of the job name.

### helo


```bash
codeplumber is the missing tool to manage your AWS CodePipeline resources.

Usage:
  codeplumber [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  load        Load a profile
  profiles    List available profiles
  run         Run codeplumber with optional tags and name filter

Flags:
  -p, --aws-profile string   Use a specific AWS profile from your AWS credential file, overwrite ENV variable AWS_PROFILE.
  -r, --aws-region string    The AWS region to use, overwrite ENV variable AWS_REGION. (default "us-east-1")
  -c, --config string        CodePlumber Configuration file location. (default "$HOME/.config/codeplumber/config.yaml")
  -d, --debug                Enable debug log, out will be saved in ./codeplumber.log
  -h, --help                 help for codeplumber
  -v, --version              version for codeplumber

Use "codeplumber [command] --help" for more information about a command.
```

### Shell competion

This project use the [Cobra](https://github.com/spf13/cobra) library, which provides a way to generate shell completion scripts.
The program have been optimized to provide profiles completions, which can be quite handy if you have many profiles defined.

Use `codeplumber completion <YOUR_SHELL>` to generate the completion script for your shell and place it in the right location.


### Limitations

There are likely many limitations. CodePipeline supports a lots of various steps, and some of them might not be supported at the moment.
This is a personal project, and I am not using all the features of CodePipeline, so I am not aware of all the edge cases.
Contributions are more than welcome! :)

## Roadmap

- [ ] Add support for CodeBuild. Similar to CodePipeline layout, but focusing on CodeBuild jobs.
- [ ] Add auto-refresh when a job is in progress.
- [ ] Provide a way to quickly navigate to non-inline CodeBuild `buildspecs` definitions.

