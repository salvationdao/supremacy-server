# Github Actions

## Build and Staging Deployment

### Required Secrets

```sh
# Private Repo Access
# Key pairs available in 1password / DevOps vault
GH_DEPLOY_LOGHELPERS_PK # Deploy private key for https://github.com/ninja-software/log_helpers
GH_DEPLOY_BRIDGE_PK     # Deploy private key for https://github.com/ninja-syndicate/supremacy-bridge
GH_DEPLOY_HUB_PK        # Deploy private key for https://github.com/ninja-syndicate/hub

# Deployment target
STAGING_SSH_HOST # The host name of the target
STAGING_SSH_KNOWN_HOSTS # ssh-keyscan -t ED25519 -p STAGING_SSH_PORT STAGING_SSH_HOST | tee github-key-temp | ssh-keygen -lf -
STAGING_SSH_PKEY
STAGING_SSH_PORT
STAGING_SSH_USER
```

### Jobs

#### Build

1. Set up private repo access using multiple deploy keys so the keys have read only access to the private repos.
    [multi key private go repos](https://gist.github.com/jrapoport/d12f60029eef017354d0ec982b918258).

    The basic steps are
    1. Create a host alias for github in `.ssh/config` with the private key file location
    2. Create the private key file
    3. Add the full repo url (ie `https://github.com/ninja-syndicate/hub`) to the git config override

2. Get the most recent version, increment by one and create a new tag, depending on branch.  
   if develop, use git hash `$PREVIOUS_TAG-dev-$GITHASH`.  
   if staging, use `$PREVIOUS_TAG-rc.1`, incrementing the `rc.` number.  
   else fail.  

   Also set any version related environment variables  
   **DEBUGGING**  
   - Error `Process completed with exit code 1.`
     Reason: Most likely on the wrong branch
     Response: This script intentionally will only tag on `develop` and `staging`
   - Error `fatal: tag 'x.y.z-rc.n' already exists`  
     Reason: There is a commit with 2 tags
     Response: Switch to staging and increment the tag on a untagged commit.
     `git switch staging && git commit --allow-empty -m 'tag error fix' && git tag x.y.z-rc.n+1 && git commit --allow-empty -m 'tag error fix'`
   - Error `fatal: No names found, cannot describe anything.`  
     Reason: There are no existing tags  
     Response: ensure that `actions/checkout@v2` has `fetch-depth: "0"`  
     Response alternitive: manually tag a commit on the same branch that the CI runs this  script on
   - Error `fatal: something something GITHASH`  
     This happens occasionally and I can't reproduce.
     Response: Contact Nathan so it can be resolved and documented.

3. Set private repo environment variables, build tools.  

4. Copy required files like configs, samples and migrations.

5. Build API server with version related environment variables.

6. Generate build info.

7. Push new tag to github.

8. Save build output to github actions.

#### Deploy

1. Wait for build to finish.

2. Retrive build output from previous job.

3. Set up ssh config using the deployment target secrets.

4. Rsync the files to `/usr/share/ninja_syndicate/gameserver_${{env.GITVERSION}}`

5. Copy the environment file from `gameserver_online` to new the new version.

6. BROKEN: backup the database

7. Update the `gameserver_online` symbolic link to the new version

8. Drop the database

9. Run migrate up

10. Restart services
