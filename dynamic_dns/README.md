# Custom Dynamic DNS

I have a few machines which have a dynamic IPs. I like to be able to resolve these dynamic IPs via a domain I control, with the minimum amount of overhead. This creates a simple executable which uses `cloudfare` API to update a DNS entry.

This can be called via cron every x minutes.
