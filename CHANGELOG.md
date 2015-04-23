# Changelog

## 0.4.1 2015-04-18

  + Add loading indicator for system information 
    request
  + bug fixes

## 0.4.0 2015-04-18

  + Add system information window for node.
    Right click node and select `system information`.
  + bug fixes

## 0.3.1 2015-04-18

  + Fix table sort - respect query order by first.
  + perf limit added to query response. Limit by 10k.
  + better query debug output for node scope queries.
  + minor UI fixes.
  + rename osquery.go to query.go

## 0.3.0 - 2015-04-16

  + fix bug with node template not showing os icon
  + when clicking a table in a node view - add select query to editor
  + Truncate node metadata in node listing.
  + Editor comments, new lines and pop last valid for current search.
  + Way better error messages for query syntax issues.
  + Add a few new default queries:
    * Third-party kernel extensions (OS X)
    * Startup items (OS X)
    * Interface information
    * Shell history
    * All users with groups
  + minor bug fixes

  - socket verbose output for errors returned from handlers

## 0.2.4 - 2015-04-16

  + better debug output for node request/time of query
  + daemon support for server
  + minor bug fixes

## 0.2.3 - 2015-04-13

  + os icon for node button
  + better query loading indicator.
  + readme updates
