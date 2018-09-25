drop table api.audit_log
;

create table api.audit_log
(
	request_id varchar(100) not null
		constraint audit_log_pkey
			primary key,
    client_id varchar(100),        
	request_timestamp timestamp,
	response_code integer,
	response_timestamp timestamp,
	duration_in_millis bigint,
	protocol varchar(20) not null,
	protocol_major integer,
	protocol_minor integer,
	request_method varchar(10) not null,
	scheme varchar(100),
	host varchar(100) not null,
	port varchar(100) not null,
	path varchar(4000),
	remote_address varchar(100),
	request_content_length bigint,
	request_header jsonb,
	request_body text,
	response_header jsonb,
	response_body text
)
;

alter table api.audit_log owner to gilcrest
;

drop function api.log_request(varchar, varchar, timestamp, integer, timestamp, bigint, varchar, integer, integer, varchar, varchar, varchar, varchar, varchar, varchar, bigint, jsonb, text, jsonb, text)
;

create function api.log_request(p_request_id character varying, p_client_id character varying, p_request_timestamp timestamp without time zone, p_response_code integer, p_response_timestamp timestamp without time zone, p_duration_in_millis bigint, p_protocol character varying, p_protocol_major integer, p_protocol_minor integer, p_request_method character varying, p_scheme character varying, p_host character varying, p_port character varying, p_path character varying, p_remote_address character varying, p_request_content_length bigint, p_request_header jsonb, p_request_body text, p_response_header jsonb, p_response_body text) returns integer
	language plpgsql
as $$
DECLARE
  v_rows_inserted INTEGER;
BEGIN
 INSERT INTO api.audit_log (request_id,
                            client_id,
                            request_timestamp,
                            response_code,
                            response_timestamp,
                            duration_in_millis,
                            protocol,
                            protocol_major,
                            protocol_minor,
                            request_method,
                            scheme,
                            host,
                            port,
                            path,
                            remote_address,
                            request_content_length,
                            request_header,
                            request_body,
                            response_header,
                            response_body
                            )
	  VALUES (p_request_id,
            p_client_id,
            p_request_timestamp,
            p_response_code,
            p_response_timestamp,
            p_duration_in_millis,
            p_protocol,
            p_protocol_major,
            p_protocol_minor,
            p_request_method,
            p_scheme,
            p_host,
            p_port,
            p_path,
            p_remote_address,
            p_request_content_length,
            p_request_header,
            p_request_body,
            p_response_header,
            p_response_body
            );
  GET DIAGNOSTICS v_rows_inserted = ROW_COUNT;
  return v_rows_inserted;
END;
$$
;

alter function api.log_request(varchar, varchar, timestamp, integer, timestamp, bigint, varchar, integer, integer, varchar, varchar, varchar, varchar, varchar, varchar, bigint, jsonb, text, jsonb, text) owner to gilcrest
;

