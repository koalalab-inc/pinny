With Software supply chain attacks on the rise, attackers are finding new ways to attack. While securing OSS depdencies is importand and imperative, some attack vectors open up because of configuration issues in GitOps workflows, in this case GitHub's workflows.

**Imposter commits, a type of Dependency Confusion**

GitHub has a lot of features that make GitOps really easy, one among them is the ability to [checkout pull requests from forks directly from the parent repo](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/reviewing-changes-in-pull-requests/checking-out-pull-requests-locally). 

GitHub shares commits between the fork and its parent - this means that forks can be created very quickly (since not all objects need to be copied over to a separate repo) and you can easily checkout and test someone's change without needing to know what their remote is.

This convenience has a trade off - when working with commit SHAs directly, how do you know if a specific commit came from the primary repo or a fork? Has the commit been reviewed and checked into the primary repo's main branch? Could you tell if the commit was authored by a legitimate maintainer of the repo?

_Imposter commits are commits that appear to be from a parent repository, but they actually belong to a fork._

This led to a notable incident in [2020 when a user created a commit in the github/dmca repo](https://github.com/github/dmca/commit/565ece486c7c1652754d7b6d2b5ed9cb4097f9d5) impersonating GitHub's then-CEO Nat Friedman, uploading a copy of what appeared to be GitHub source code.

GitHub now and displays a warning in the UI above any commits that don't belong to a branch in the parent repository, for example:

![Screenshot 2024-04-11 at 12 43 26 PM](https://github.com/koalalab-inc/pinny/assets/149300820/0f5b9293-4718-4a64-80c5-0c023d7980b1)

But there aren't similar protections when using GitHub from a CLI or through the API.

This is specifically problematic when using GitHub Actions in CI/CD.

**Impostor commits in CI/CD**

GitHub fails to distinguish between fork and non-fork SHA references, forks can bypass security settings on GitHub Actions that would otherwise restrict actions to only “trusted” sources (such as GitHub themselves or the repository’s own organization).
GitHub has added the practice of checking for forked vs parent branch when using Actions with SHA commits in their [best practices blog.](https://docs.github.com/en/actions/learn-github-actions/finding-and-customizing-actions#using-shas)


Refrences:
1. [Chainguard's Impostor Commit Blog](https://www.chainguard.dev/unchained/what-the-fork-imposter-commits-in-github-actions-and-ci-cd)
2. [GitHub Actions could be so uch better](https://blog.yossarian.net/2023/09/22/GitHub-Actions-could-be-so-much-better)
