Quick and dirty hack to make nice with bitbucket privately hosted repositories that only allow ssh access.  In order to set this up,

    go get github.com/myENA/bbgoget
    
Alternately you can download the rpm or the tarball from the Releases tab.

As an example configuration, with Apache httpd running on 443 proxying for bitbucket, you could add the following to
httpd.conf section before the redirects to bitbucket's http service:

            RewriteEngine on
            RewriteCond "%{QUERY_STRING}" "go-get"
            RewriteRule "^/(.*)" "http://localhost:8800/$1" [P]

This assumes bbgoget running locally on port 8800.

The integration with bitbucket is minimal / dumb.  It has no knowledge whether or not you actually have a repository 
at the requested location.  It doesn't really need to, go get will figure out soon enough if the response ends up
going nowhere.  The up side is this makes configuration easier.  No API access required, minimal configuration.
As an added benefit of being unconcerned with tight integration with bitbucket, there is a very good chance this
would work with other git servers.  Also - and most importantly - this made it very easy to author.


We've only tested it with our specific configuration.