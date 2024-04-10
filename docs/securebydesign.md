![HERO - SecurebyDesign - 1600x600px](https://github.com/koalalab-inc/pinny/assets/149300820/48df3261-7bc7-4c41-8436-9412cde25261)
[CISA,](https://www.cisa.gov/) the US government's Cyber Defense Agency, has been pushing for adoption of [Secure-by-design principles](https://www.cisa.gov/securebydesign) by software manufacturers. 
Google has also published a guide to understand the [High-Level-Principles of Software development which would lead to creation of "Secure-By-Design" software](https://storage.googleapis.com/gweb-research2023-media/pubtools/pdf/6f28d2ea12b39c0f7b7c220b8fcc0f89db91e5a9.pdf).

We can summarise these principles into 3 broad buckets:
  1. **_Invariants_** of the Software system:
  _Defined secure defaults which are enforced through systems settings and toolings to always remain true, even when system is under attack._
     
     Examples:

     a. Every request to all methods of a web services API is mediated by a well-defined authentication and authorization policy.
     
     b. For all call sites of a SQL query, the query string is solely composed of trustworthy snippets and all untrustworthy parameters as value sbound to query parameters.
     
     c. All network traffic is protected using secure protocol such as TLS.

     > NOTE:While it's easy to define _Invariants_ but actually what is required is requisite tooling baked into the Software product and SDLC to ensure that these invariants hold true.

3. Designing for understandability and assurance aka _"Developers are users too"_

   **_Security Posture of a software system is evolving_**, any new update and configuration change changes the security posture. Most security mishaps occur because the users developing or using the security systems can make mistakes.
   So while defining secure defaults or invariants of the system is imperative but the enforcement can only happen with tooling developed keeping the users in mind and having a user-centric design of the SDLC ecosystem.

4. Secure developer ecosystems to provide assurance at scale

    The complexity of the software ecosystem begets the use and creation of: 
  
      a. _Safe Coding_: Usage of secure libraries and components, memory-safe languages, secure IDEs etc

      b. _Safe Deployment_: Creation of secure default configs of all deployment environments, **_hardended build environments_** etc

      c. _Developer tooling for all application archetypes_: Tooling to apply secure defaults or invariants across all application archetypes


   
