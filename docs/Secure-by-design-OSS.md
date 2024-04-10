Software supply chain is all the components and processes that go into the creation of software. In Modern software, the software supply chain would be: 

1. Components
   
     a. Proprietary Code
   
     b. Open-Source Code Dependencies
   
3. Processes

     Source-code to Production Process Going through the CI/CD or Devops Pipelines.

![Software-Supply-Chain-Snyk](https://github.com/koalalab-inc/pinny/assets/149300820/c3c7d9c2-b203-4c6f-a255-1fb963b521bb)


Software industry stands on the contributions of open source developers. [Linux foundation estimates that Free and Open Source Software(FOSS) consitutes **70-90%** of all modern software](https://www.linuxfoundation.org/blog/blog/a-summary-of-census-ii-open-source-software-application-libraries-the-world-depends-on) being written right now. 
Just as an example, [Sonatype projects Java(Maven) component request volume](https://www.sonatype.com/state-of-the-software-supply-chain/open-source-supply-and-demand) in 2023 to be about 1 trillion with 25% growth Y-o-Y.

**_Secure-By-Design in OSS Dependencies?_**

While ["Secure-By-Design" Principles](https://github.com/koalalab-inc/pinny/blob/main/docs/securebydesign.md) maybe hard to apply when using OSS Dependencies but a few ways to achieve that would be:
1. Using secure base images( [Alpine Linux](https://www.alpinelinux.org/) for example)
2. Using Package Managers for OSS dependencies wherever available.
   _Package Managers themselves maybe open to attack(as seen in the [XZ attack](https://www.techrepublic.com/article/xz-backdoor-linux/)) but mostly have policies in place to produce secure software itself._

Still, there are many OSS dependencies which get directly used/downloaded from the internet, such dependencies include ```dockerfiles, docker-compose files, GitHub Actions``` etc.
Such dependencies are open to attack vectors like:

1. [TypoSquatting](https://owasp.org/www-project-top-10-ci-cd-security-risks/CICD-SEC-03-Dependency-Chain-Abuse): Publication of malicious packages with similar names to those of popular packages in the hope that a developer will misspell a package name and unintentionally fetch the typosquatted package.
2. [Dependency confusion](https://owasp.org/www-project-top-10-ci-cd-security-risks/CICD-SEC-03-Dependency-Chain-Abuse): Publication of malicious packages in public repositories with the same name as internal package names, in an attempt to trick clients into downloading the malicious package rather than the private one.
3. [Brandjacking](https://owasp.org/www-project-top-10-ci-cd-security-risks/CICD-SEC-03-Dependency-Chain-Abuse): Publication of malicious packages in a manner that is consistent with the naming convention or other characteristics of a specific brand’s package, in an attempt to get unsuspecting developers to fetch these packages due to falsely associating them with the trusted brand.
4. [RepoJacking](https://github.blog/2024-02-21-how-to-stay-safe-from-repo-jacking/): A type of attack where the original publisher changes the name of their organization/repository and an attacker takes over that name to publish malicious packages.
5. [Dependency Hijacking](https://owasp.org/www-project-top-10-ci-cd-security-risks/CICD-SEC-03-Dependency-Chain-Abuse): Obtaining control of the account of a package maintainer on the public repository, in order to upload a new, malicious version of a widely used package, with the intent of compromising unsuspecting clients who pull the latest version of the package.

   While all these attack vectors are possible in package manager managed OSS dependencies but the risk is more pronounced when directly using from the internet.

These threat vectors can easily be solved by a practice of hash-pinning the image/version of the OSS dependency that the code is using. [Hash-pinning is a recommended security practice](https://docs.github.com/en/actions/security-guides/security-hardening-for-github-actions#using-third-party-actions) and is [effective against many common types of attacks](https://www.paloaltonetworks.com/blog/prisma-cloud/unpinnable-actions-github-security/).
Lot of security professionals have also recommended a similar practice, examples [here](https://blog.rafaelgss.dev/why-you-should-pin-actions-by-commit-hash), [here](https://medium.com/ochrona/preventing-dependency-confusion-attacks-in-python-fa6058ac972f) and [here](https://michaelheap.com/ensure-github-actions-pinned-sha/).

[**Automated Hash-Pinning**](https://github.com/koalalab-inc/pinny) is a tool that can help in the _**"secure-by-design"**_ journey, specially when using non-package-manager managed OSS dependencies.





References:
1. [Anatomy of a software supply chain attack](https://www.technologydecisions.com.au/content/security/article/anatomy-of-a-supply-chain-software-attack-440028396)
2. [State of software supply chain by Sonatype](https://www.sonatype.com/state-of-the-software-supply-chain/open-source-supply-and-demand)
3. [OWASP CI/CD Top 10's CI/CD SEC-3: Dependency Chain Abuse](https://owasp.org/www-project-top-10-ci-cd-security-risks/CICD-SEC-03-Dependency-Chain-Abuse)
![Screenshot 2024-04-11 at 2 10 36 AM](https://github.com/koalalab-inc/pinny/assets/149300820/afcc5dd8-c214-4a3a-be37-6191276811fd)
![Screenshot 2024-04-11 at 2 10 26 AM](https://github.com/koalalab-inc/pinny/assets/149300820/49fa7169-84fe-473c-ba0e-5425a81b11d3)
![Screenshot 2024-04-11 at 2 10 13 AM](https://github.com/koalalab-inc/pinny/assets/149300820/1e96d166-f58c-4c58-a70d-bef13442f07e)
