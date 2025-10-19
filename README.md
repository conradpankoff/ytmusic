# ytmusic

This application maintains a local cache of videos from YouTube to allow for
offline use on devices where there is no official YouTube support.

## Basic Schema Architecture

1. The main data models hold data about resources on YouTube (channels,
   videos, playlists, and playlist contents). See "main data model" section of
   `schema.sql`.
2. Some data (specifically that which is fetched from YouTube like titles,
   descriptions, upload dates, etc) may not be available at the time of record
   creation, so it can be null, but is expected to be populated soon after
   record creation. This data, when structurally relevant (e.g. some
   identifiers), is maintained across table via triggers. See "asynchronous
   relation maintenance triggers" section of `schema.sql`.
3. Views are defined to combine data from multiple tables via joins so that
   the application can use simpler queries, and so that search indexes can
   be built. See "view definitions" section of `schema.sql`.
4. Indexes are built to provide full text search over the views mentioned
   above to power the list/search functions in the application and should
   expose all the data necessary to render a list for whatever data it
   represents (e.g. for a list of videos, it would include some channel
   metadata like title and id). See "index definitions" section of `schema.sql`.
5. The search indexes are maintained via triggers. See "index triggers"
   section of `schema.sql`.
6. The search indexes can be recomputed via `insert into $x_search (...)
   select ... from $x_search_view` queries. See "index population" section of
   `schema.sql`.

This architecture and its goals should be maintained.

## Implementation Notes

* Use `sorm` to access the database to ensure the correct type mappings occur
  between the database and the application.
* Minimise the use of manually constructed SQL in application code, preferring
  the use of `sqlbuilder` via `qsorm`.
* Maintain consistency of implementation across similar functionality.
  * Only extract something out into a separate package if it's used three or
    more times and moving it out of context does not make it hard to
    understand. Exception for enumerated values (e.g. the `queuenames`
    package) - these should always be extracted if used in more than one
    place.
