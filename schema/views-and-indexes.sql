begin;

drop trigger if exists channels__update_search_on_insert;
drop trigger if exists channels__update_search_on_update;
drop trigger if exists channels__update_search_on_delete;
drop trigger if exists playlists__update_search_on_insert;
drop trigger if exists playlists__update_search_on_update;
drop trigger if exists playlists__update_search_on_delete;
drop trigger if exists videos__update_search_on_insert;
drop trigger if exists videos__update_search_on_update;
drop trigger if exists videos__update_search_on_delete;

drop table if exists channel_search;
drop table if exists playlist_search;
drop table if exists video_search;

drop view if exists channel_search_view;
drop view if exists playlist_search_view;
drop view if exists video_search_view;
drop view if exists video_in_playlist_view;

-- views for search/list pages

create view channel_search_view as select
  c.id as channel_id,
  c.created_at as channel_created_at,
  c.external_id as channel_external_id,
  c.title as channel_title,
  c.metadata_updated_at as channel_metadata_updated_at,
  c.thumbnail_updated_at as channel_thumbnail_updated_at
from channels c;

create view playlist_search_view as select
  c.id as channel_id,
  c.created_at as channel_created_at,
  coalesce(c.external_id, p.channel_external_id) as channel_external_id,
  coalesce(c.title, '') as channel_title,
  c.metadata_updated_at as channel_metadata_updated_at,
  c.thumbnail_updated_at as channel_thumbnail_updated_at,
  p.id as playlist_id,
  p.created_at as playlist_created_at,
  p.external_id as playlist_external_id,
  p.title as playlist_title,
  p.metadata_updated_at as playlist_metadata_updated_at,
  p.thumbnail_updated_at as playlist_thumbnail_updated_at
from playlists p
left join channels c
  on c.id = p.channel_id or c.external_id = p.channel_external_id;

create view video_search_view as select
  c.id as channel_id,
  c.created_at as channel_created_at,
  coalesce(c.external_id, v.channel_external_id) as channel_external_id,
  coalesce(c.title, '') as channel_title,
  c.metadata_updated_at as channel_metadata_updated_at,
  c.thumbnail_updated_at as channel_thumbnail_updated_at,
  v.id as video_id,
  v.created_at as video_created_at,
  v.external_id as video_external_id,
  v.title as video_title,
  v.description as video_description,
  v.metadata_updated_at as video_metadata_updated_at,
  v.thumbnail_updated_at as video_thumbnail_updated_at,
  v.downloaded_at as video_downloaded_at,
  v.transcoded_360_at as video_transcoded_360_at,
  v.transcoded_720_at as video_transcoded_720_at,
  v.audio_extracted_at as video_audio_extracted_at
from videos v
left join channels c
  on c.id = v.channel_id or c.external_id = v.channel_external_id;

create view video_in_playlist_view as select
  c.id as channel_id,
  c.created_at as channel_created_at,
  coalesce(c.external_id, '') as channel_external_id,
  coalesce(c.title, '') as channel_title,
  c.metadata_updated_at as channel_metadata_updated_at,
  c.thumbnail_updated_at as channel_thumbnail_updated_at,
  p.id as playlist_id,
  p.created_at as playlist_created_at,
  coalesce(p.external_id, pv.playlist_external_id) as playlist_external_id,
  coalesce(p.title, '') as playlist_title,
  p.metadata_updated_at as playlist_metadata_updated_at,
  p.thumbnail_updated_at as playlist_thumbnail_updated_at,
  pv.id as playlist_video_id,
  pv.created_at as playlist_video_created_at,
  pv.position as playlist_video_position,
  v.id as video_id,
  v.created_at as video_created_at,
  coalesce(v.external_id, pv.video_external_id) as video_external_id,
  coalesce(v.title, '') as video_title,
  coalesce(v.description, '') as video_description,
  v.metadata_updated_at as video_metadata_updated_at,
  v.thumbnail_updated_at as video_thumbnail_updated_at,
  v.downloaded_at as video_downloaded_at,
  v.transcoded_360_at as video_transcoded_360_at,
  v.transcoded_720_at as video_transcoded_720_at,
  v.audio_extracted_at as video_audio_extracted_at
from playlist_videos pv
left join playlists p
  on p.id = pv.playlist_id or p.external_id = pv.playlist_external_id
left join videos v
  on v.id = pv.video_id or v.external_id = pv.video_external_id
left join channels c
  on c.id = v.channel_id or c.external_id = v.channel_external_id;

-- indexes for search pages

create virtual table channel_search using fts5(
  content='channel_search_view', content_rowid='channel_id',
  channel_id unindexed, channel_created_at unindexed, channel_external_id,
  channel_title,
  channel_metadata_updated_at unindexed, channel_thumbnail_updated_at unindexed
);

create virtual table playlist_search using fts5(
  content='playlist_search_view', content_rowid='playlist_id',
  channel_id unindexed, channel_created_at unindexed, channel_external_id,
  channel_title,
  channel_metadata_updated_at unindexed, channel_thumbnail_updated_at unindexed,
  playlist_id unindexed, playlist_created_at unindexed, playlist_external_id,
  playlist_title,
  playlist_metadata_updated_at unindexed, playlist_thumbnail_updated_at unindexed
);

create virtual table video_search using fts5(
  content='video_search_view', content_rowid='video_id',
  channel_id unindexed, channel_created_at unindexed, channel_external_id,
  channel_title,
  channel_metadata_updated_at unindexed, channel_thumbnail_updated_at unindexed,
  video_id unindexed, video_created_at unindexed, video_external_id,
  video_title, video_description,
  video_metadata_updated_at unindexed, video_thumbnail_updated_at unindexed, video_downloaded_at unindexed, video_transcoded_360_at unindexed, video_transcoded_720_at unindexed, video_audio_extracted_at unindexed
);

-- keep the search indexes updated when the source data changes

create trigger channels__update_search_on_insert after insert on channels
begin
  insert into channel_search (rowid, channel_external_id, channel_title)
    select
      channel_id,
      channel_external_id, channel_title
    from channel_search_view
    where channel_id = new.id;

  update playlist_search set channel_title = new.title where channel_external_id = new.external_id;
  update video_search set channel_title = new.title where channel_external_id = new.external_id;
end;

create trigger channels__update_search_on_update after update of external_id, title on channels
begin
  update channel_search
    set
      channel_external_id = new.external_id, channel_title = new.title
    where rowid = new.id;

  update playlist_search set channel_title = new.title where channel_external_id = new.external_id;
  update video_search set channel_title = new.title where channel_external_id = new.external_id;
end;

create trigger channels__update_search_on_delete after delete on playlists
begin
  insert into channel_search (channel_search, rowid) values ('delete', old.id);
end;

create trigger playlists__update_search_on_insert after insert on playlists
begin
  insert into playlist_search (rowid, playlist_external_id, playlist_title, channel_external_id, channel_title)
    select
      playlist_id, playlist_external_id, playlist_title,
      channel_external_id, channel_title
    from playlist_search_view
    where playlist_id = new.id;
end;

create trigger playlists__update_search_on_update after update of external_id, title, description on playlists
begin
  update playlist_search
    set
      playlist_external_id = new.external_id, playlist_title = new.title,
      channel_external_id = new.channel_external_id
    where rowid = new.id;
end;

create trigger playlists__update_search_on_delete after delete on playlists
begin
  insert into playlist_search (playlist_search, rowid) values ('delete', old.id);
end;

create trigger videos__update_search_on_insert after insert on videos
begin
  insert into video_search (rowid, video_external_id, video_title, video_description, channel_external_id, channel_title)
    select
      video_id, video_external_id, video_title, video_description,
      channel_external_id, channel_title
    from video_search_view
    where video_id = new.id;
end;

create trigger videos__update_search_on_update after update of external_id, title, description on videos
begin
  update video_search
    set
      video_external_id = new.external_id, video_title = new.title, video_description = new.description,
      channel_external_id = new.channel_external_id
    where rowid = new.id;
end;

create trigger videos__update_search_on_delete after delete on videos
begin
  insert into video_search (video_search, rowid) values ('delete', old.id);
end;

-- populate indexes and test queries

insert into channel_search (rowid, channel_external_id, channel_title)
  select
    channel_id,
    channel_external_id, channel_title
  from channel_search_view;

insert into playlist_search (rowid, playlist_external_id, playlist_title, channel_external_id, channel_title)
  select
    playlist_id,
    playlist_external_id, playlist_title,
    channel_external_id, channel_title
  from playlist_search_view;

insert into video_search (rowid, video_external_id, video_title, video_description, channel_external_id, channel_title)
  select
    video_id,
    video_external_id, video_title, video_description,
    channel_external_id, channel_title
  from video_search_view;

select 'channels';
select
  channel_id, channel_title
from channel_search where channel_search match 'grymm';

select 'videos';
select
  video_id, video_title,
  channel_id, channel_title
from video_search where video_search match 'grymm';

select 'playlists';
select
  playlist_id, playlist_title,
  channel_id, channel_title
from playlist_search where playlist_search match 'grymm';

rollback;
