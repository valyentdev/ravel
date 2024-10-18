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

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: fleets; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.fleets (
    id text NOT NULL,
    name text NOT NULL,
    namespace text NOT NULL,
    created_at timestamp without time zone NOT NULL
);


--
-- Name: gateways; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.gateways (
    id text NOT NULL,
    name text NOT NULL,
    fleet_id text NOT NULL,
    namespace text NOT NULL,
    target_port integer NOT NULL,
    protocol text NOT NULL
);


--
-- Name: instance_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.instance_events (
    id text NOT NULL,
    type text NOT NULL,
    origin text NOT NULL,
    payload jsonb NOT NULL,
    instance_id text NOT NULL,
    machine_id text NOT NULL,
    status text NOT NULL,
    "timestamp" timestamp without time zone NOT NULL
);


--
-- Name: machine_versions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.machine_versions (
    id text NOT NULL,
    machine_id text NOT NULL,
    config jsonb NOT NULL
);


--
-- Name: machines; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.machines (
    id text NOT NULL,
    node text NOT NULL,
    namespace text NOT NULL,
    fleet_id text NOT NULL,
    instance_id text NOT NULL,
    region text NOT NULL,
    created_at timestamp without time zone DEFAULT timezone('utc'::text, now()) NOT NULL,
    updated_at timestamp without time zone DEFAULT timezone('utc'::text, now()) NOT NULL,
    destroyed boolean DEFAULT false NOT NULL,
    machine_version text NOT NULL
);


--
-- Name: namespaces; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.namespaces (
    name text NOT NULL,
    created_at timestamp without time zone DEFAULT timezone('utc'::text, now()) NOT NULL
);


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schema_migrations (
    version character varying(255) NOT NULL
);


--
-- Name: fleets fleets_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fleets
    ADD CONSTRAINT fleets_pkey PRIMARY KEY (id);


--
-- Name: gateways gateways_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.gateways
    ADD CONSTRAINT gateways_pkey PRIMARY KEY (id);


--
-- Name: instance_events instance_events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.instance_events
    ADD CONSTRAINT instance_events_pkey PRIMARY KEY (id);


--
-- Name: machine_versions machine_versions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.machine_versions
    ADD CONSTRAINT machine_versions_pkey PRIMARY KEY (id);


--
-- Name: machines machines_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.machines
    ADD CONSTRAINT machines_pkey PRIMARY KEY (id);


--
-- Name: namespaces namespaces_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.namespaces
    ADD CONSTRAINT namespaces_pkey PRIMARY KEY (name);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: fleets unique_name_in_namespace; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fleets
    ADD CONSTRAINT unique_name_in_namespace UNIQUE (name, namespace);


--
-- Name: fleets_namespace_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX fleets_namespace_idx ON public.fleets USING btree (namespace);


--
-- Name: fleets_namespace_name_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX fleets_namespace_name_idx ON public.fleets USING btree (namespace, name);


--
-- Name: gateways_name_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX gateways_name_idx ON public.gateways USING btree (namespace, name);


--
-- Name: instance_events_machine_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX instance_events_machine_id_idx ON public.instance_events USING btree (machine_id);


--
-- Name: fleets fleets_namespace_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fleets
    ADD CONSTRAINT fleets_namespace_fkey FOREIGN KEY (namespace) REFERENCES public.namespaces(name);


--
-- Name: gateways gateways_fleet_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.gateways
    ADD CONSTRAINT gateways_fleet_id_fkey FOREIGN KEY (fleet_id) REFERENCES public.fleets(id);


--
-- Name: gateways gateways_namespace_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.gateways
    ADD CONSTRAINT gateways_namespace_fkey FOREIGN KEY (namespace) REFERENCES public.namespaces(name);


--
-- Name: instance_events instance_events_machine_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.instance_events
    ADD CONSTRAINT instance_events_machine_id_fkey FOREIGN KEY (machine_id) REFERENCES public.machines(id) ON DELETE CASCADE;


--
-- Name: machines machines_fleet_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.machines
    ADD CONSTRAINT machines_fleet_id_fkey FOREIGN KEY (fleet_id) REFERENCES public.fleets(id);


--
-- Name: machines machines_namespace_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.machines
    ADD CONSTRAINT machines_namespace_fkey FOREIGN KEY (namespace) REFERENCES public.namespaces(name);


--
-- PostgreSQL database dump complete
--


--
-- Dbmate schema migrations
--

INSERT INTO public.schema_migrations (version) VALUES
    ('20240809111759'),
    ('20241005214512'),
    ('20241018074516');
