// Copyright 2017 Openprovider Authors. All rights reserved.
// Use of this source code is governed by a license
// that can be found in the LICENSE file.

package config

var usage = `
whoisd - Whois Daemon

Usage:
  whoisd install | remove | start | stop | status | cacheFlush | cacheStatus
  whoisd [ -t | --test ] [ -option | -option ... ]
  whoisd -h | --help
  whoisd -v | --version

Commands:
  install           Install as service (is only valid for Linux and Mac Os X)
  remove            Remove service
  start             Start service
  stop              Stop service
  status            Check service status
  cacheFlush        Removes all items from memory cache
  cacheStatus       List all items in memory cache in file

  -h --help         Show this screen
  -v --version      Show version
  -t --test         Test mode

Options:
  -config=<path>        Path to config file (used in /etc/whoisd/whoisd.conf)
  -mapping=<path>       Path to mapping file (used in /etc/whoisd/conf.d/mapping.json)
  -host=<host/IP>       Host name or IP address
  -port=<port>          Port number
  -work=<number>        Number of active workers (default 1000)
  -conn=<number>        Number of active connections (default 1000)
  -storage=<type>       Type of storage (Elasticsearch, Mysql or Dummy for testing)
  -shost=<host/IP>      Storage host name or IP address
  -sport=<port>         Storage port number
  -suser=<name>         User name for storage login
  -spassword=<***>      Password of storage user
  -base=<name>          Storage index or database name
  -table=<name>         Storage type or table name
  -cache=<bool>         Enable/disable memory cache
  -cexpiration=<number> Set expiration of items in cache in minutes
  -ccleanup=<number>    Sets interval for deletion of expired items from cache
`

// Usage - get usage information
func Usage() string {
	return usage
}
