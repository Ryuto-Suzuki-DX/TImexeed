--
-- PostgreSQL database dump
--

\restrict OB8Ha7gXpoVjBK1AFpskvypzUEUBp0TH0Qk2hEgXPKCtPvEjG4eBLc8IvfQXvLH

-- Dumped from database version 16.14
-- Dumped by pg_dump version 16.14

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: public; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA public;


--
-- Name: SCHEMA public; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON SCHEMA public IS 'standard public schema';


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: api_operation_logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.api_operation_logs (
    id bigint NOT NULL,
    user_id bigint,
    email character varying(255),
    role character varying(50),
    method character varying(10) NOT NULL,
    path character varying(255) NOT NULL,
    status_code bigint NOT NULL,
    client_ip character varying(100),
    user_agent text,
    duration_ms bigint NOT NULL,
    error_message text,
    started_at timestamp with time zone NOT NULL,
    finished_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


--
-- Name: api_operation_logs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.api_operation_logs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: api_operation_logs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.api_operation_logs_id_seq OWNED BY public.api_operation_logs.id;


--
-- Name: attendance_breaks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.attendance_breaks (
    id bigint NOT NULL,
    attendance_day_id bigint NOT NULL,
    break_start_at timestamp with time zone NOT NULL,
    break_end_at timestamp with time zone NOT NULL,
    break_memo text,
    is_deleted boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


--
-- Name: attendance_breaks_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.attendance_breaks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: attendance_breaks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.attendance_breaks_id_seq OWNED BY public.attendance_breaks.id;


--
-- Name: attendance_days; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.attendance_days (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    work_date date NOT NULL,
    plan_attendance_type_id bigint NOT NULL,
    actual_work_status character varying(30) DEFAULT 'NORMAL'::character varying NOT NULL,
    plan_start_at timestamp with time zone,
    plan_end_at timestamp with time zone,
    actual_start_at timestamp with time zone,
    actual_end_at timestamp with time zone,
    scheduled_work_minutes bigint,
    remote_work_allowance_flag boolean DEFAULT false NOT NULL,
    is_deleted boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


--
-- Name: attendance_days_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.attendance_days_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: attendance_days_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.attendance_days_id_seq OWNED BY public.attendance_days.id;


--
-- Name: attendance_realtime_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.attendance_realtime_events (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    event_date date NOT NULL,
    event_type character varying(30) NOT NULL,
    event_at timestamp with time zone NOT NULL,
    note text,
    client_ip character varying(100),
    user_agent text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


--
-- Name: attendance_realtime_events_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.attendance_realtime_events_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: attendance_realtime_events_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.attendance_realtime_events_id_seq OWNED BY public.attendance_realtime_events.id;


--
-- Name: attendance_transport_expenses; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.attendance_transport_expenses (
    id bigint NOT NULL,
    attendance_day_id bigint NOT NULL,
    sort_order bigint DEFAULT 0 NOT NULL,
    transport_from character varying(100) NOT NULL,
    transport_to character varying(100) NOT NULL,
    transport_method character varying(50) NOT NULL,
    transport_amount bigint DEFAULT 0 NOT NULL,
    transport_memo character varying(255),
    is_deleted boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


--
-- Name: attendance_transport_expenses_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.attendance_transport_expenses_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: attendance_transport_expenses_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.attendance_transport_expenses_id_seq OWNED BY public.attendance_transport_expenses.id;


--
-- Name: attendance_types; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.attendance_types (
    id bigint NOT NULL,
    code character varying(50) NOT NULL,
    name character varying(100) NOT NULL,
    category character varying(50) NOT NULL,
    is_worked boolean DEFAULT false NOT NULL,
    requires_request boolean DEFAULT false NOT NULL,
    sync_plan_actual boolean DEFAULT false NOT NULL,
    allow_actual_time_input boolean DEFAULT true NOT NULL,
    allow_break_input boolean DEFAULT true NOT NULL,
    allow_transport_input boolean DEFAULT true NOT NULL,
    allow_late_flag boolean DEFAULT true NOT NULL,
    allow_early_leave_flag boolean DEFAULT true NOT NULL,
    allow_absence_flag boolean DEFAULT true NOT NULL,
    allow_sick_leave_flag boolean DEFAULT true NOT NULL,
    display_order bigint DEFAULT 0 NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    is_deleted boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


--
-- Name: attendance_types_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.attendance_types_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: attendance_types_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.attendance_types_id_seq OWNED BY public.attendance_types.id;


--
-- Name: departments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.departments (
    id bigint NOT NULL,
    name character varying(100) NOT NULL,
    is_deleted boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


--
-- Name: departments_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.departments_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: departments_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.departments_id_seq OWNED BY public.departments.id;


--
-- Name: expenses; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.expenses (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    target_month date NOT NULL,
    expense_date date NOT NULL,
    amount bigint NOT NULL,
    description text NOT NULL,
    memo text,
    original_file_name character varying(255),
    stored_file_name character varying(255),
    file_url text,
    drive_file_id character varying(255),
    external_storage_link_id bigint,
    mime_type character varying(100),
    size_bytes bigint,
    is_deleted boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


--
-- Name: expenses_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.expenses_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: expenses_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.expenses_id_seq OWNED BY public.expenses.id;


--
-- Name: external_storage_links; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.external_storage_links (
    id bigint NOT NULL,
    link_type character varying(80) NOT NULL,
    link_name character varying(100) NOT NULL,
    url text NOT NULL,
    description text,
    memo text,
    is_deleted boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


--
-- Name: external_storage_links_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.external_storage_links_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: external_storage_links_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.external_storage_links_id_seq OWNED BY public.external_storage_links.id;


--
-- Name: holiday_dates; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.holiday_dates (
    id bigint NOT NULL,
    holiday_date date NOT NULL,
    holiday_name character varying(100) NOT NULL,
    is_deleted boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


--
-- Name: holiday_dates_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.holiday_dates_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: holiday_dates_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.holiday_dates_id_seq OWNED BY public.holiday_dates.id;


--
-- Name: monthly_attendance_requests; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.monthly_attendance_requests (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    target_year bigint NOT NULL,
    target_month bigint NOT NULL,
    status character varying(30) NOT NULL,
    request_memo text,
    requested_at timestamp with time zone,
    approved_by bigint,
    approved_at timestamp with time zone,
    rejected_reason text,
    rejected_at timestamp with time zone,
    canceled_reason text,
    canceled_at timestamp with time zone,
    is_deleted boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


--
-- Name: monthly_attendance_requests_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.monthly_attendance_requests_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: monthly_attendance_requests_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.monthly_attendance_requests_id_seq OWNED BY public.monthly_attendance_requests.id;


--
-- Name: monthly_commuter_passes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.monthly_commuter_passes (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    target_year bigint NOT NULL,
    target_month bigint NOT NULL,
    commuter_from character varying(100),
    commuter_to character varying(100),
    commuter_method character varying(50),
    commuter_amount bigint,
    is_deleted boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


--
-- Name: monthly_commuter_passes_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.monthly_commuter_passes_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: monthly_commuter_passes_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.monthly_commuter_passes_id_seq OWNED BY public.monthly_commuter_passes.id;


--
-- Name: notification_reminders; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notification_reminders (
    id bigint NOT NULL,
    title character varying(150) NOT NULL,
    message text NOT NULL,
    day_offset_from_month_end bigint DEFAULT 0 NOT NULL,
    send_hour bigint DEFAULT 9 NOT NULL,
    send_minute bigint DEFAULT 0 NOT NULL,
    is_enabled boolean DEFAULT true NOT NULL,
    is_deleted boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


--
-- Name: notification_reminders_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.notification_reminders_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: notification_reminders_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.notification_reminders_id_seq OWNED BY public.notification_reminders.id;


--
-- Name: notifications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications (
    id bigint NOT NULL,
    notification_group_id character varying(36),
    user_id bigint NOT NULL,
    title character varying(150) NOT NULL,
    message text NOT NULL,
    is_read boolean DEFAULT false NOT NULL,
    read_at timestamp with time zone,
    is_deleted boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


--
-- Name: notifications_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.notifications_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: notifications_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.notifications_id_seq OWNED BY public.notifications.id;


--
-- Name: paid_leave_usages; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.paid_leave_usages (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    usage_date date NOT NULL,
    usage_days numeric(4,1) NOT NULL,
    is_manual boolean DEFAULT false NOT NULL,
    memo character varying(255),
    is_deleted boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


--
-- Name: paid_leave_usages_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.paid_leave_usages_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: paid_leave_usages_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.paid_leave_usages_id_seq OWNED BY public.paid_leave_usages.id;


--
-- Name: personal_information_drive_folders; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.personal_information_drive_folders (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    external_storage_link_id bigint NOT NULL,
    folder_name character varying(255) NOT NULL,
    drive_folder_id character varying(255) NOT NULL,
    folder_url text NOT NULL,
    synced_at timestamp with time zone,
    is_deleted boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


--
-- Name: personal_information_drive_folders_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.personal_information_drive_folders_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: personal_information_drive_folders_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.personal_information_drive_folders_id_seq OWNED BY public.personal_information_drive_folders.id;


--
-- Name: shared_document_drive_folders; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.shared_document_drive_folders (
    id bigint NOT NULL,
    folder_name text NOT NULL,
    description text,
    drive_folder_id text NOT NULL,
    folder_url text NOT NULL,
    synced_at timestamp with time zone,
    is_deleted boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


--
-- Name: shared_document_drive_folders_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.shared_document_drive_folders_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: shared_document_drive_folders_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.shared_document_drive_folders_id_seq OWNED BY public.shared_document_drive_folders.id;


--
-- Name: user_salary_details; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_salary_details (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    salary_type character varying(20) NOT NULL,
    base_amount bigint DEFAULT 0 NOT NULL,
    extra_allowance_amount bigint DEFAULT 0 NOT NULL,
    extra_allowance_memo text,
    fixed_deduction_amount bigint DEFAULT 0 NOT NULL,
    fixed_deduction_memo text,
    is_payroll_target boolean DEFAULT true NOT NULL,
    effective_from date NOT NULL,
    effective_to date,
    memo text,
    is_deleted boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


--
-- Name: user_salary_details_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.user_salary_details_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: user_salary_details_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.user_salary_details_id_seq OWNED BY public.user_salary_details.id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id bigint NOT NULL,
    name character varying(100) NOT NULL,
    email character varying(255) NOT NULL,
    password_hash character varying(255) NOT NULL,
    role character varying(20) DEFAULT 'USER'::character varying NOT NULL,
    department_id bigint,
    hire_date date NOT NULL,
    retirement_date date,
    must_change_password boolean DEFAULT false NOT NULL,
    password_changed_at timestamp with time zone,
    is_deleted boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: api_operation_logs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.api_operation_logs ALTER COLUMN id SET DEFAULT nextval('public.api_operation_logs_id_seq'::regclass);


--
-- Name: attendance_breaks id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attendance_breaks ALTER COLUMN id SET DEFAULT nextval('public.attendance_breaks_id_seq'::regclass);


--
-- Name: attendance_days id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attendance_days ALTER COLUMN id SET DEFAULT nextval('public.attendance_days_id_seq'::regclass);


--
-- Name: attendance_realtime_events id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attendance_realtime_events ALTER COLUMN id SET DEFAULT nextval('public.attendance_realtime_events_id_seq'::regclass);


--
-- Name: attendance_transport_expenses id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attendance_transport_expenses ALTER COLUMN id SET DEFAULT nextval('public.attendance_transport_expenses_id_seq'::regclass);


--
-- Name: attendance_types id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attendance_types ALTER COLUMN id SET DEFAULT nextval('public.attendance_types_id_seq'::regclass);


--
-- Name: departments id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.departments ALTER COLUMN id SET DEFAULT nextval('public.departments_id_seq'::regclass);


--
-- Name: expenses id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.expenses ALTER COLUMN id SET DEFAULT nextval('public.expenses_id_seq'::regclass);


--
-- Name: external_storage_links id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.external_storage_links ALTER COLUMN id SET DEFAULT nextval('public.external_storage_links_id_seq'::regclass);


--
-- Name: holiday_dates id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.holiday_dates ALTER COLUMN id SET DEFAULT nextval('public.holiday_dates_id_seq'::regclass);


--
-- Name: monthly_attendance_requests id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.monthly_attendance_requests ALTER COLUMN id SET DEFAULT nextval('public.monthly_attendance_requests_id_seq'::regclass);


--
-- Name: monthly_commuter_passes id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.monthly_commuter_passes ALTER COLUMN id SET DEFAULT nextval('public.monthly_commuter_passes_id_seq'::regclass);


--
-- Name: notification_reminders id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notification_reminders ALTER COLUMN id SET DEFAULT nextval('public.notification_reminders_id_seq'::regclass);


--
-- Name: notifications id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications ALTER COLUMN id SET DEFAULT nextval('public.notifications_id_seq'::regclass);


--
-- Name: paid_leave_usages id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.paid_leave_usages ALTER COLUMN id SET DEFAULT nextval('public.paid_leave_usages_id_seq'::regclass);


--
-- Name: personal_information_drive_folders id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.personal_information_drive_folders ALTER COLUMN id SET DEFAULT nextval('public.personal_information_drive_folders_id_seq'::regclass);


--
-- Name: shared_document_drive_folders id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.shared_document_drive_folders ALTER COLUMN id SET DEFAULT nextval('public.shared_document_drive_folders_id_seq'::regclass);


--
-- Name: user_salary_details id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_salary_details ALTER COLUMN id SET DEFAULT nextval('public.user_salary_details_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Name: api_operation_logs api_operation_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.api_operation_logs
    ADD CONSTRAINT api_operation_logs_pkey PRIMARY KEY (id);


--
-- Name: attendance_breaks attendance_breaks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attendance_breaks
    ADD CONSTRAINT attendance_breaks_pkey PRIMARY KEY (id);


--
-- Name: attendance_days attendance_days_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attendance_days
    ADD CONSTRAINT attendance_days_pkey PRIMARY KEY (id);


--
-- Name: attendance_realtime_events attendance_realtime_events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attendance_realtime_events
    ADD CONSTRAINT attendance_realtime_events_pkey PRIMARY KEY (id);


--
-- Name: attendance_transport_expenses attendance_transport_expenses_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attendance_transport_expenses
    ADD CONSTRAINT attendance_transport_expenses_pkey PRIMARY KEY (id);


--
-- Name: attendance_types attendance_types_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attendance_types
    ADD CONSTRAINT attendance_types_pkey PRIMARY KEY (id);


--
-- Name: departments departments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.departments
    ADD CONSTRAINT departments_pkey PRIMARY KEY (id);


--
-- Name: expenses expenses_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.expenses
    ADD CONSTRAINT expenses_pkey PRIMARY KEY (id);


--
-- Name: external_storage_links external_storage_links_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.external_storage_links
    ADD CONSTRAINT external_storage_links_pkey PRIMARY KEY (id);


--
-- Name: holiday_dates holiday_dates_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.holiday_dates
    ADD CONSTRAINT holiday_dates_pkey PRIMARY KEY (id);


--
-- Name: monthly_attendance_requests monthly_attendance_requests_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.monthly_attendance_requests
    ADD CONSTRAINT monthly_attendance_requests_pkey PRIMARY KEY (id);


--
-- Name: monthly_commuter_passes monthly_commuter_passes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.monthly_commuter_passes
    ADD CONSTRAINT monthly_commuter_passes_pkey PRIMARY KEY (id);


--
-- Name: notification_reminders notification_reminders_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notification_reminders
    ADD CONSTRAINT notification_reminders_pkey PRIMARY KEY (id);


--
-- Name: notifications notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (id);


--
-- Name: paid_leave_usages paid_leave_usages_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.paid_leave_usages
    ADD CONSTRAINT paid_leave_usages_pkey PRIMARY KEY (id);


--
-- Name: personal_information_drive_folders personal_information_drive_folders_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.personal_information_drive_folders
    ADD CONSTRAINT personal_information_drive_folders_pkey PRIMARY KEY (id);


--
-- Name: shared_document_drive_folders shared_document_drive_folders_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.shared_document_drive_folders
    ADD CONSTRAINT shared_document_drive_folders_pkey PRIMARY KEY (id);


--
-- Name: user_salary_details user_salary_details_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_salary_details
    ADD CONSTRAINT user_salary_details_pkey PRIMARY KEY (id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: idx_api_operation_logs_method; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_api_operation_logs_method ON public.api_operation_logs USING btree (method);


--
-- Name: idx_api_operation_logs_path; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_api_operation_logs_path ON public.api_operation_logs USING btree (path);


--
-- Name: idx_api_operation_logs_started_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_api_operation_logs_started_at ON public.api_operation_logs USING btree (started_at);


--
-- Name: idx_api_operation_logs_status_code; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_api_operation_logs_status_code ON public.api_operation_logs USING btree (status_code);


--
-- Name: idx_api_operation_logs_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_api_operation_logs_user_id ON public.api_operation_logs USING btree (user_id);


--
-- Name: idx_attendance_breaks_attendance_day_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_attendance_breaks_attendance_day_id ON public.attendance_breaks USING btree (attendance_day_id);


--
-- Name: idx_attendance_days_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_attendance_days_user_id ON public.attendance_days USING btree (user_id);


--
-- Name: idx_attendance_days_work_date; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_attendance_days_work_date ON public.attendance_days USING btree (work_date);


--
-- Name: idx_attendance_realtime_event_unique; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_attendance_realtime_event_unique ON public.attendance_realtime_events USING btree (user_id, event_date, event_type);


--
-- Name: idx_attendance_realtime_events_event_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_attendance_realtime_events_event_at ON public.attendance_realtime_events USING btree (event_at);


--
-- Name: idx_attendance_realtime_events_event_date; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_attendance_realtime_events_event_date ON public.attendance_realtime_events USING btree (event_date);


--
-- Name: idx_attendance_realtime_events_event_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_attendance_realtime_events_event_type ON public.attendance_realtime_events USING btree (event_type);


--
-- Name: idx_attendance_realtime_events_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_attendance_realtime_events_user_id ON public.attendance_realtime_events USING btree (user_id);


--
-- Name: idx_attendance_transport_expenses_attendance_day_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_attendance_transport_expenses_attendance_day_id ON public.attendance_transport_expenses USING btree (attendance_day_id);


--
-- Name: idx_attendance_transport_expenses_is_deleted; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_attendance_transport_expenses_is_deleted ON public.attendance_transport_expenses USING btree (is_deleted);


--
-- Name: idx_attendance_types_code; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_attendance_types_code ON public.attendance_types USING btree (code);


--
-- Name: idx_departments_name; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_departments_name ON public.departments USING btree (name);


--
-- Name: idx_expenses_drive_file_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_expenses_drive_file_id ON public.expenses USING btree (drive_file_id);


--
-- Name: idx_expenses_expense_date; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_expenses_expense_date ON public.expenses USING btree (expense_date);


--
-- Name: idx_expenses_external_storage_link_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_expenses_external_storage_link_id ON public.expenses USING btree (external_storage_link_id);


--
-- Name: idx_expenses_target_month; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_expenses_target_month ON public.expenses USING btree (target_month);


--
-- Name: idx_expenses_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_expenses_user_id ON public.expenses USING btree (user_id);


--
-- Name: idx_external_storage_links_link_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_external_storage_links_link_type ON public.external_storage_links USING btree (link_type);


--
-- Name: idx_holiday_dates_holiday_date; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_holiday_dates_holiday_date ON public.holiday_dates USING btree (holiday_date);


--
-- Name: idx_monthly_attendance_requests_target_month; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_monthly_attendance_requests_target_month ON public.monthly_attendance_requests USING btree (target_month);


--
-- Name: idx_monthly_attendance_requests_target_year; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_monthly_attendance_requests_target_year ON public.monthly_attendance_requests USING btree (target_year);


--
-- Name: idx_monthly_attendance_requests_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_monthly_attendance_requests_user_id ON public.monthly_attendance_requests USING btree (user_id);


--
-- Name: idx_monthly_commuter_passes_target_month; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_monthly_commuter_passes_target_month ON public.monthly_commuter_passes USING btree (target_month);


--
-- Name: idx_monthly_commuter_passes_target_year; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_monthly_commuter_passes_target_year ON public.monthly_commuter_passes USING btree (target_year);


--
-- Name: idx_monthly_commuter_passes_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_monthly_commuter_passes_user_id ON public.monthly_commuter_passes USING btree (user_id);


--
-- Name: idx_notifications_notification_group_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_notifications_notification_group_id ON public.notifications USING btree (notification_group_id);


--
-- Name: idx_notifications_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_notifications_user_id ON public.notifications USING btree (user_id);


--
-- Name: idx_paid_leave_usages_usage_date; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_paid_leave_usages_usage_date ON public.paid_leave_usages USING btree (usage_date);


--
-- Name: idx_paid_leave_usages_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_paid_leave_usages_user_id ON public.paid_leave_usages USING btree (user_id);


--
-- Name: idx_personal_information_drive_folders_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_personal_information_drive_folders_deleted_at ON public.personal_information_drive_folders USING btree (deleted_at);


--
-- Name: idx_personal_information_drive_folders_drive_folder_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_personal_information_drive_folders_drive_folder_id ON public.personal_information_drive_folders USING btree (drive_folder_id);


--
-- Name: idx_personal_information_drive_folders_external_storage_link_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_personal_information_drive_folders_external_storage_link_id ON public.personal_information_drive_folders USING btree (external_storage_link_id);


--
-- Name: idx_personal_information_drive_folders_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_personal_information_drive_folders_user_id ON public.personal_information_drive_folders USING btree (user_id);


--
-- Name: idx_shared_document_drive_folders_drive_folder_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_shared_document_drive_folders_drive_folder_id ON public.shared_document_drive_folders USING btree (drive_folder_id);


--
-- Name: idx_user_salary_details_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_salary_details_deleted_at ON public.user_salary_details USING btree (deleted_at);


--
-- Name: idx_user_salary_details_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_salary_details_user_id ON public.user_salary_details USING btree (user_id);


--
-- Name: idx_users_email; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_users_email ON public.users USING btree (email);


--
-- Name: attendance_breaks fk_attendance_breaks_attendance_day; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attendance_breaks
    ADD CONSTRAINT fk_attendance_breaks_attendance_day FOREIGN KEY (attendance_day_id) REFERENCES public.attendance_days(id);


--
-- Name: attendance_days fk_attendance_days_plan_attendance_type; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attendance_days
    ADD CONSTRAINT fk_attendance_days_plan_attendance_type FOREIGN KEY (plan_attendance_type_id) REFERENCES public.attendance_types(id);


--
-- Name: attendance_realtime_events fk_attendance_realtime_events_user; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attendance_realtime_events
    ADD CONSTRAINT fk_attendance_realtime_events_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: attendance_transport_expenses fk_attendance_transport_expenses_attendance_day; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attendance_transport_expenses
    ADD CONSTRAINT fk_attendance_transport_expenses_attendance_day FOREIGN KEY (attendance_day_id) REFERENCES public.attendance_days(id);


--
-- Name: expenses fk_expenses_user; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.expenses
    ADD CONSTRAINT fk_expenses_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: paid_leave_usages fk_paid_leave_usages_user; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.paid_leave_usages
    ADD CONSTRAINT fk_paid_leave_usages_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- PostgreSQL database dump complete
--

\unrestrict OB8Ha7gXpoVjBK1AFpskvypzUEUBp0TH0Qk2hEgXPKCtPvEjG4eBLc8IvfQXvLH

