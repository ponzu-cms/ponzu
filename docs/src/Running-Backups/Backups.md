title: Running Backups on Ponzu systems

Both the databases `system.db` & `analytics.db`, and the `/uploads` directory can be backed up over HTTP using `wget`, `curl`, etc. All of which are located at the `/admin/backup` route and require HTTP Basic Auth. In order to enable backups, you must add a user/password pair inside the CMS Configuration at `/admin/configure` near the bottom of the page.

All backups are made using a `GET` request to the `/admin/backup` path with a query parameter of `?source={system,analytics,uploads}` (only one source can be included in the URL).

Here are some full backup scripts to use or modify to fit your needs:
[https://github.com/ponzu-cms/backup-scripts](https://github.com/ponzu-cms/backup-scripts)

## System & Analytics
The `system.db` & `analytics.db` data files are sent uncompressed in their original form as they exist on your server. No temporary copy is stored on the origin server, and it is possible that the backup could fail so checking for successful backups is recommended. See https://github.com/boltdb/bolt#database-backups for more information about how BoltDB handles HTTP backups.

An example backup request for the `system.db` data file would look like:
```bash
$ curl --user user:pass "https://example.com/admin/backup?source=system" > system.db.bak
```

## Uploads
The `/uploads` directory is gzip compressed and archived as a tar file, stored in the temporary directory (typically `/tmp` on Linux) on your origin server with a timestamp in the file name. 

An example backup request for the `/uploads` directory would look like:
```bash
$ curl --user user:pass "https://example.com/admin/backup?source=uploads" > uploads.tar.gz
# unarchive the tarball with gzip 
$ tar xzf uploads.tar.gz
```