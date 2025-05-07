update videos set thumbnail_updated_at = null, transcoded_360_at = null, transcoded_720_at = null, audio_extracted_at = null;

insert into jobs (created_at, queue_name, payload, run_after, failure_delay, attempts_remaining, error_messages, output_messages)
  select current_timestamp, 'video_download', external_id, current_timestamp, 5000000000, 5, json_array(), json_array() from videos where downloaded_at is null;

insert into jobs (created_at, queue_name, payload, run_after, failure_delay, attempts_remaining, error_messages, output_messages)
  select current_timestamp, 'video_update_thumbnail', external_id, current_timestamp, 5000000000, 5, json_array(), json_array() from videos where downloaded_at is not null and thumbnail_updated_at is null;

insert into jobs (created_at, queue_name, payload, run_after, failure_delay, attempts_remaining, error_messages, output_messages)
  select current_timestamp, 'video_transcode', external_id || '?size=360', current_timestamp, 5000000000, 5, json_array(), json_array() from videos where downloaded_at is not null and transcoded_360_at is null;

insert into jobs (created_at, queue_name, payload, run_after, failure_delay, attempts_remaining, error_messages, output_messages)
  select current_timestamp, 'video_transcode', external_id || '?size=720', current_timestamp, 5000000000, 5, json_array(), json_array() from videos where downloaded_at is not null and transcoded_720_at is null;

insert into jobs (created_at, queue_name, payload, run_after, failure_delay, attempts_remaining, error_messages, output_messages)
  select current_timestamp, 'video_extract_audio', external_id, current_timestamp, 5000000000, 5, json_array(), json_array() from videos where downloaded_at is not null and audio_extracted_at is null;
