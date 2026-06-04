/*
 Navicat Premium Dump SQL

 Source Server         : 153.35.238.82_25432
 Source Server Type    : PostgreSQL
 Source Server Version : 150015 (150015)
 Source Host           : 153.35.238.82:25432
 Source Catalog        : chatbot
 Source Schema         : public

 Target Server Type    : PostgreSQL
 Target Server Version : 150015 (150015)
 File Encoding         : 65001

 Date: 21/05/2026 22:49:36
*/


-- ----------------------------
-- Table structure for biz_announcement
-- ----------------------------
DROP TABLE IF EXISTS "public"."biz_announcement";
CREATE TABLE "public"."biz_announcement" (
  "id" varchar(64) COLLATE "pg_catalog"."default" NOT NULL,
  "title" varchar(512) COLLATE "pg_catalog"."default" NOT NULL,
  "content" varchar(8000) COLLATE "pg_catalog"."default" NOT NULL DEFAULT ''::character varying,
  "sort_no" int4 NOT NULL DEFAULT 0,
  "status" int4 NOT NULL DEFAULT 1,
  "publish_time" timestamp(6),
  "create_time" timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "update_time" timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted" int4 NOT NULL DEFAULT 0
)
;
ALTER TABLE "public"."biz_announcement" OWNER TO "postgres";
COMMENT ON TABLE "public"."biz_announcement" IS '公告';

-- ----------------------------
-- Records of biz_announcement
-- ----------------------------
BEGIN;
INSERT INTO "public"."biz_announcement" ("id", "title", "content", "sort_no", "status", "publish_time", "create_time", "update_time", "deleted") VALUES ('00000000-0000-0000-0000-000000000040', '欢迎使用 ChatBot 管理后台', '您可在首页查看套餐与调用额度，前往「套餐购买」可体验模拟下单流程。', 1, 1, '2026-05-15 09:31:46.823377', '2026-05-15 09:31:46.823377', '2026-05-15 09:31:46.823377', 0);
COMMIT;

-- ----------------------------
-- Table structure for biz_package
-- ----------------------------
DROP TABLE IF EXISTS "public"."biz_package";
CREATE TABLE "public"."biz_package" (
  "id" varchar(64) COLLATE "pg_catalog"."default" NOT NULL,
  "code" varchar(64) COLLATE "pg_catalog"."default" NOT NULL,
  "name" varchar(256) COLLATE "pg_catalog"."default" NOT NULL,
  "description" varchar(2000) COLLATE "pg_catalog"."default" NOT NULL DEFAULT ''::character varying,
  "quota_calls" int8 NOT NULL DEFAULT 0,
  "sort_no" int4 NOT NULL DEFAULT 0,
  "status" int4 NOT NULL DEFAULT 1,
  "create_time" timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "update_time" timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted" int4 NOT NULL DEFAULT 0
)
;
ALTER TABLE "public"."biz_package" OWNER TO "postgres";
COMMENT ON COLUMN "public"."biz_package"."quota_calls" IS '套餐包含的可调用次数';
COMMENT ON TABLE "public"."biz_package" IS '套餐定义';

-- ----------------------------
-- Records of biz_package
-- ----------------------------
BEGIN;
INSERT INTO "public"."biz_package" ("id", "code", "name", "description", "quota_calls", "sort_no", "status", "create_time", "update_time", "deleted") VALUES ('00000000-0000-0000-0000-000000000020', 'trial', '体验包', '新用户体验链路，少量调用次数', 20, 10, 1, '2026-05-15 09:31:46.88172', '2026-05-15 09:31:46.88172', 0);
INSERT INTO "public"."biz_package" ("id", "code", "name", "description", "quota_calls", "sort_no", "status", "create_time", "update_time", "deleted") VALUES ('00000000-0000-0000-0000-000000000021', 'standard', '标准包', '适合日常试用与轻量业务', 500, 20, 1, '2026-05-15 09:31:46.88172', '2026-05-15 09:31:46.88172', 0);
INSERT INTO "public"."biz_package" ("id", "code", "name", "description", "quota_calls", "sort_no", "status", "create_time", "update_time", "deleted") VALUES ('00000000-0000-0000-0000-000000000022', 'pro', '专业包', '更高调用额度', 5000, 30, 1, '2026-05-15 09:31:46.88172', '2026-05-15 09:31:46.88172', 0);
COMMIT;

-- ----------------------------
-- Table structure for biz_usage_log
-- ----------------------------
DROP TABLE IF EXISTS "public"."biz_usage_log";
CREATE TABLE "public"."biz_usage_log" (
  "id" varchar(64) COLLATE "pg_catalog"."default" NOT NULL,
  "user_id" varchar(64) COLLATE "pg_catalog"."default" NOT NULL,
  "channel" varchar(64) COLLATE "pg_catalog"."default" NOT NULL DEFAULT 'experience'::character varying,
  "ref_text" varchar(512) COLLATE "pg_catalog"."default" NOT NULL DEFAULT ''::character varying,
  "create_time" timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
)
;
ALTER TABLE "public"."biz_usage_log" OWNER TO "postgres";
COMMENT ON TABLE "public"."biz_usage_log" IS '调用流水';

-- ----------------------------
-- Records of biz_usage_log
-- ----------------------------
BEGIN;
COMMIT;

-- ----------------------------
-- Table structure for biz_user_preference
-- ----------------------------
DROP TABLE IF EXISTS "public"."biz_user_preference";
CREATE TABLE "public"."biz_user_preference" (
  "user_id" varchar(64) COLLATE "pg_catalog"."default" NOT NULL,
  "standard_text" varchar(8000) COLLATE "pg_catalog"."default" NOT NULL DEFAULT ''::character varying,
  "create_time" timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "update_time" timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
)
;
ALTER TABLE "public"."biz_user_preference" OWNER TO "postgres";
COMMENT ON COLUMN "public"."biz_user_preference"."user_id" IS '用户 ID';
COMMENT ON COLUMN "public"."biz_user_preference"."standard_text" IS '录音时展示的标准文本';
COMMENT ON TABLE "public"."biz_user_preference" IS '用户偏好（标准朗读文本等）';

-- ----------------------------
-- Records of biz_user_preference
-- ----------------------------
BEGIN;
COMMIT;

-- ----------------------------
-- Table structure for biz_user_quota
-- ----------------------------
DROP TABLE IF EXISTS "public"."biz_user_quota";
CREATE TABLE "public"."biz_user_quota" (
  "user_id" varchar(64) COLLATE "pg_catalog"."default" NOT NULL,
  "package_id" varchar(64) COLLATE "pg_catalog"."default" NOT NULL DEFAULT ''::character varying,
  "package_title" varchar(256) COLLATE "pg_catalog"."default" NOT NULL DEFAULT ''::character varying,
  "total_quota" int8 NOT NULL DEFAULT 0,
  "used_quota" int8 NOT NULL DEFAULT 0,
  "update_time" timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
)
;
ALTER TABLE "public"."biz_user_quota" OWNER TO "postgres";
COMMENT ON COLUMN "public"."biz_user_quota"."package_title" IS '展示用套餐名称快照';
COMMENT ON TABLE "public"."biz_user_quota" IS '用户额度';

-- ----------------------------
-- Records of biz_user_quota
-- ----------------------------
BEGIN;
INSERT INTO "public"."biz_user_quota" ("user_id", "package_id", "package_title", "total_quota", "used_quota", "update_time") VALUES ('0c3f344d-a617-47f7-b75d-fae33a4a62e2', '00000000-0000-0000-0000-000000000020', '体验包', 20, 0, '2026-05-15 09:55:55.812616');
COMMIT;

-- ----------------------------
-- Table structure for biz_user_voice
-- ----------------------------
DROP TABLE IF EXISTS "public"."biz_user_voice";
CREATE TABLE "public"."biz_user_voice" (
  "id" varchar(64) COLLATE "pg_catalog"."default" NOT NULL,
  "user_id" varchar(64) COLLATE "pg_catalog"."default" NOT NULL,
  "file_name" varchar(256) COLLATE "pg_catalog"."default" NOT NULL,
  "upload_time" timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "duration_seconds" float8 NOT NULL DEFAULT 0,
  "is_active" int4 NOT NULL DEFAULT 0,
  "create_time" timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "update_time" timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted" int4 NOT NULL DEFAULT 0
)
;
ALTER TABLE "public"."biz_user_voice" OWNER TO "postgres";
COMMENT ON COLUMN "public"."biz_user_voice"."file_name" IS '相对于 audio_target_path 的文件名，如 {uuid}.wav';
COMMENT ON COLUMN "public"."biz_user_voice"."duration_seconds" IS '音频时长（秒）';
COMMENT ON COLUMN "public"."biz_user_voice"."is_active" IS '是否当前激活：1 是 0 否';
COMMENT ON TABLE "public"."biz_user_voice" IS '用户语音样本';

-- ----------------------------
-- Records of biz_user_voice
-- ----------------------------
BEGIN;
INSERT INTO "public"."biz_user_voice" ("id", "user_id", "file_name", "upload_time", "duration_seconds", "is_active", "create_time", "update_time", "deleted") VALUES ('e95579f9-914d-48db-8c83-293fdd52804e', '0c3f344d-a617-47f7-b75d-fae33a4a62e2', 'e95579f9-914d-48db-8c83-293fdd52804e.wav', '2026-05-21 13:53:14.494529', 15.325167, 1, '2026-05-21 13:53:14.494529', '2026-05-21 13:53:19.751574', 0);
INSERT INTO "public"."biz_user_voice" ("id", "user_id", "file_name", "upload_time", "duration_seconds", "is_active", "create_time", "update_time", "deleted") VALUES ('d37155c4-280c-464f-b397-404afebc6404', '0c3f344d-a617-47f7-b75d-fae33a4a62e2', 'd37155c4-280c-464f-b397-404afebc6404.wav', '2026-05-21 15:39:33.953319', 27.446, 0, '2026-05-21 15:39:33.953319', '2026-05-21 15:39:33.953319', 0);
INSERT INTO "public"."biz_user_voice" ("id", "user_id", "file_name", "upload_time", "duration_seconds", "is_active", "create_time", "update_time", "deleted") VALUES ('45737ccb-6b13-42ad-8b40-84ba46dd66d3', '00000000-0000-0000-0000-000000000001', '45737ccb-6b13-42ad-8b40-84ba46dd66d3.wav', '2026-05-21 16:32:28.549422', 11.911833, 0, '2026-05-21 16:32:28.549422', '2026-05-21 16:34:26.131954', 0);
INSERT INTO "public"."biz_user_voice" ("id", "user_id", "file_name", "upload_time", "duration_seconds", "is_active", "create_time", "update_time", "deleted") VALUES ('9bc54fbc-cb1f-4ee7-aff2-deda5f9a4a1f', '00000000-0000-0000-0000-000000000001', '9bc54fbc-cb1f-4ee7-aff2-deda5f9a4a1f.wav', '2026-05-21 16:34:24.355452', 14.675, 1, '2026-05-21 16:34:24.355452', '2026-05-21 16:34:26.131954', 0);
INSERT INTO "public"."biz_user_voice" ("id", "user_id", "file_name", "upload_time", "duration_seconds", "is_active", "create_time", "update_time", "deleted") VALUES ('cd392d85-2f16-4697-9eeb-0527eb1c1bf0', '00000000-0000-0000-0000-000000000001', 'cd392d85-2f16-4697-9eeb-0527eb1c1bf0.wav', '2026-05-21 16:33:20.578139', 0.301875, 0, '2026-05-21 16:33:20.578139', '2026-05-21 16:34:26.131954', 0);
INSERT INTO "public"."biz_user_voice" ("id", "user_id", "file_name", "upload_time", "duration_seconds", "is_active", "create_time", "update_time", "deleted") VALUES ('969f6c67-5fbe-42c7-80b0-cf022d49691a', '00000000-0000-0000-0000-000000000001', '969f6c67-5fbe-42c7-80b0-cf022d49691a.wav', '2026-05-15 15:36:44.900149', 3.29725, 0, '2026-05-15 15:36:44.900149', '2026-05-21 16:34:26.131954', 0);
INSERT INTO "public"."biz_user_voice" ("id", "user_id", "file_name", "upload_time", "duration_seconds", "is_active", "create_time", "update_time", "deleted") VALUES ('f2d4fc17-0626-4473-bef7-638cf75d4db7', '00000000-0000-0000-0000-000000000001', 'f2d4fc17-0626-4473-bef7-638cf75d4db7.wav', '2026-05-21 16:32:13.132604', 3.715208, 0, '2026-05-21 16:32:13.132604', '2026-05-21 16:34:26.131954', 0);
INSERT INTO "public"."biz_user_voice" ("id", "user_id", "file_name", "upload_time", "duration_seconds", "is_active", "create_time", "update_time", "deleted") VALUES ('36b92c9b-a502-4d1d-b1e3-87b5892fe81d', '00000000-0000-0000-0000-000000000001', '36b92c9b-a502-4d1d-b1e3-87b5892fe81d.wav', '2026-05-15 09:53:23.019559', 1.671833, 0, '2026-05-15 09:53:23.019559', '2026-05-21 16:34:26.131954', 0);
INSERT INTO "public"."biz_user_voice" ("id", "user_id", "file_name", "upload_time", "duration_seconds", "is_active", "create_time", "update_time", "deleted") VALUES ('f2bc9273-173d-4c66-b0ce-a72acb2cdfe7', '0c3f344d-a617-47f7-b75d-fae33a4a62e2', 'f2bc9273-173d-4c66-b0ce-a72acb2cdfe7.wav', '2026-05-15 10:04:19.25566', 5.735333, 0, '2026-05-15 10:04:19.25566', '2026-05-21 13:53:19.751574', 0);
INSERT INTO "public"."biz_user_voice" ("id", "user_id", "file_name", "upload_time", "duration_seconds", "is_active", "create_time", "update_time", "deleted") VALUES ('1c280ebd-d8bb-4032-81af-47d0b81a5835', '0c3f344d-a617-47f7-b75d-fae33a4a62e2', '1c280ebd-d8bb-4032-81af-47d0b81a5835.wav', '2026-05-21 13:52:39.272178', 0.464417, 0, '2026-05-21 13:52:39.272178', '2026-05-21 13:53:19.751574', 0);
INSERT INTO "public"."biz_user_voice" ("id", "user_id", "file_name", "upload_time", "duration_seconds", "is_active", "create_time", "update_time", "deleted") VALUES ('d4806490-7e49-46f9-b748-5db4291b95e8', '00000000-0000-0000-0000-000000000001', 'd4806490-7e49-46f9-b748-5db4291b95e8.wav', '2026-05-15 09:54:08.945211', 2.322, 0, '2026-05-15 09:54:08.945211', '2026-05-21 16:34:26.131954', 0);
INSERT INTO "public"."biz_user_voice" ("id", "user_id", "file_name", "upload_time", "duration_seconds", "is_active", "create_time", "update_time", "deleted") VALUES ('d47a7c15-ab1c-44e1-8116-6826c64a6221', '00000000-0000-0000-0000-000000000001', 'd47a7c15-ab1c-44e1-8116-6826c64a6221.wav', '2026-05-21 15:46:52.328929', 3.552667, 0, '2026-05-21 15:46:52.328929', '2026-05-21 16:34:26.131954', 0);
INSERT INTO "public"."biz_user_voice" ("id", "user_id", "file_name", "upload_time", "duration_seconds", "is_active", "create_time", "update_time", "deleted") VALUES ('5ec1f6f0-448e-4bbf-8914-ec4110fdc1bc', '00000000-0000-0000-0000-000000000001', '5ec1f6f0-448e-4bbf-8914-ec4110fdc1bc.wav', '2026-05-21 15:47:01.39079', 3.134708, 0, '2026-05-21 15:47:01.39079', '2026-05-21 16:34:26.131954', 0);
INSERT INTO "public"."biz_user_voice" ("id", "user_id", "file_name", "upload_time", "duration_seconds", "is_active", "create_time", "update_time", "deleted") VALUES ('5ca6cf2b-0d36-4394-83c3-ecd06067812d', '00000000-0000-0000-0000-000000000001', '5ca6cf2b-0d36-4394-83c3-ecd06067812d.wav', '2026-05-15 10:13:10.621985', 15.743125, 0, '2026-05-15 10:13:10.621985', '2026-05-21 16:34:26.131954', 0);
INSERT INTO "public"."biz_user_voice" ("id", "user_id", "file_name", "upload_time", "duration_seconds", "is_active", "create_time", "update_time", "deleted") VALUES ('476e6656-e73f-46a9-934a-074bf6ff4a18', '00000000-0000-0000-0000-000000000001', '476e6656-e73f-46a9-934a-074bf6ff4a18.wav', '2026-05-15 15:31:48.420313', 15.162625, 0, '2026-05-15 15:31:48.420313', '2026-05-21 16:34:26.131954', 0);
INSERT INTO "public"."biz_user_voice" ("id", "user_id", "file_name", "upload_time", "duration_seconds", "is_active", "create_time", "update_time", "deleted") VALUES ('a287fafc-79af-47f1-9c11-895a70e2f358', '00000000-0000-0000-0000-000000000001', 'a287fafc-79af-47f1-9c11-895a70e2f358.wav', '2026-05-15 15:36:55.503668', 6.710583, 0, '2026-05-15 15:36:55.503668', '2026-05-21 16:34:26.131954', 0);
INSERT INTO "public"."biz_user_voice" ("id", "user_id", "file_name", "upload_time", "duration_seconds", "is_active", "create_time", "update_time", "deleted") VALUES ('1ca7cad2-b0e1-4092-a920-905710197c33', '00000000-0000-0000-0000-000000000001', '1ca7cad2-b0e1-4092-a920-905710197c33.wav', '2026-05-21 15:47:16.407752', 11.099125, 0, '2026-05-21 15:47:16.407752', '2026-05-21 16:34:26.131954', 0);
INSERT INTO "public"."biz_user_voice" ("id", "user_id", "file_name", "upload_time", "duration_seconds", "is_active", "create_time", "update_time", "deleted") VALUES ('9776ceb3-dbc6-4a65-aa36-2054064bf534', '00000000-0000-0000-0000-000000000001', '9776ceb3-dbc6-4a65-aa36-2054064bf534.wav', '2026-05-21 16:32:00.374504', 1.764708, 0, '2026-05-21 16:32:00.374504', '2026-05-21 16:34:26.131954', 0);
COMMIT;

-- ----------------------------
-- Table structure for sys_role
-- ----------------------------
DROP TABLE IF EXISTS "public"."sys_role";
CREATE TABLE "public"."sys_role" (
  "id" varchar(64) COLLATE "pg_catalog"."default" NOT NULL,
  "code" varchar(64) COLLATE "pg_catalog"."default" NOT NULL,
  "name" varchar(128) COLLATE "pg_catalog"."default" NOT NULL,
  "status" int4 NOT NULL DEFAULT 1,
  "create_time" timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "update_time" timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted" int4 NOT NULL DEFAULT 0
)
;
ALTER TABLE "public"."sys_role" OWNER TO "postgres";
COMMENT ON COLUMN "public"."sys_role"."id" IS '主键';
COMMENT ON COLUMN "public"."sys_role"."code" IS '角色编码（唯一）';
COMMENT ON COLUMN "public"."sys_role"."name" IS '角色名称';
COMMENT ON COLUMN "public"."sys_role"."status" IS '状态：1启用 0禁用';
COMMENT ON COLUMN "public"."sys_role"."create_time" IS '创建时间';
COMMENT ON COLUMN "public"."sys_role"."update_time" IS '更新时间';
COMMENT ON COLUMN "public"."sys_role"."deleted" IS '逻辑删除：0否 1是';
COMMENT ON TABLE "public"."sys_role" IS '系统角色';

-- ----------------------------
-- Records of sys_role
-- ----------------------------
BEGIN;
INSERT INTO "public"."sys_role" ("id", "code", "name", "status", "create_time", "update_time", "deleted") VALUES ('00000000-0000-0000-0000-000000000002', 'admin', '管理员', 1, '2026-05-15 09:31:47.131452', '2026-05-15 09:31:47.131452', 0);
INSERT INTO "public"."sys_role" ("id", "code", "name", "status", "create_time", "update_time", "deleted") VALUES ('00000000-0000-0000-0000-000000000003', 'member', '普通用户', 1, '2026-05-15 09:31:47.131452', '2026-05-15 09:31:47.131452', 0);
COMMIT;

-- ----------------------------
-- Table structure for sys_role_route
-- ----------------------------
DROP TABLE IF EXISTS "public"."sys_role_route";
CREATE TABLE "public"."sys_role_route" (
  "id" varchar(128) COLLATE "pg_catalog"."default" NOT NULL,
  "role_id" varchar(64) COLLATE "pg_catalog"."default" NOT NULL,
  "route_id" varchar(64) COLLATE "pg_catalog"."default" NOT NULL,
  "create_time" timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "update_time" timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted" int4 NOT NULL DEFAULT 0
)
;
ALTER TABLE "public"."sys_role_route" OWNER TO "postgres";
COMMENT ON TABLE "public"."sys_role_route" IS '角色-路由授权';

-- ----------------------------
-- Records of sys_role_route
-- ----------------------------
BEGIN;
INSERT INTO "public"."sys_role_route" ("id", "role_id", "route_id", "create_time", "update_time", "deleted") VALUES ('00000000-0000-0000-0000-000000000037', '00000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000010', '2026-05-15 09:31:47.178752', '2026-05-15 09:31:47.178752', 0);
INSERT INTO "public"."sys_role_route" ("id", "role_id", "route_id", "create_time", "update_time", "deleted") VALUES ('00000000-0000-0000-0000-000000000038', '00000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000014', '2026-05-15 09:31:47.178752', '2026-05-15 09:31:47.178752', 0);
INSERT INTO "public"."sys_role_route" ("id", "role_id", "route_id", "create_time", "update_time", "deleted") VALUES ('8f6c28f6-2ca6-4afa-8c5f-1b2accf46c31', '00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000010', '2026-05-21 15:05:45.540804', '2026-05-21 15:05:45.540804', 0);
INSERT INTO "public"."sys_role_route" ("id", "role_id", "route_id", "create_time", "update_time", "deleted") VALUES ('3573c7f9-015b-4ab2-ac9d-ba5947be45db', '00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000011', '2026-05-21 15:05:45.540804', '2026-05-21 15:05:45.540804', 0);
INSERT INTO "public"."sys_role_route" ("id", "role_id", "route_id", "create_time", "update_time", "deleted") VALUES ('411f7ed5-7ddb-4184-a87c-f7de58747722', '00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000012', '2026-05-21 15:05:45.540804', '2026-05-21 15:05:45.540804', 0);
INSERT INTO "public"."sys_role_route" ("id", "role_id", "route_id", "create_time", "update_time", "deleted") VALUES ('4c9f4153-132d-4630-be4a-8b28696ee734', '00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000013', '2026-05-21 15:05:45.540804', '2026-05-21 15:05:45.540804', 0);
COMMIT;

-- ----------------------------
-- Table structure for sys_route
-- ----------------------------
DROP TABLE IF EXISTS "public"."sys_route";
CREATE TABLE "public"."sys_route" (
  "id" varchar(64) COLLATE "pg_catalog"."default" NOT NULL,
  "parent_id" varchar(64) COLLATE "pg_catalog"."default" NOT NULL DEFAULT ''::character varying,
  "name" varchar(128) COLLATE "pg_catalog"."default" NOT NULL,
  "path" varchar(512) COLLATE "pg_catalog"."default" NOT NULL,
  "component" varchar(512) COLLATE "pg_catalog"."default" NOT NULL,
  "title" varchar(256) COLLATE "pg_catalog"."default" NOT NULL DEFAULT ''::character varying,
  "i18n_key" varchar(256) COLLATE "pg_catalog"."default" NOT NULL DEFAULT ''::character varying,
  "icon" varchar(256) COLLATE "pg_catalog"."default" NOT NULL DEFAULT ''::character varying,
  "order_no" int4 NOT NULL DEFAULT 0,
  "hide_in_menu" int4 NOT NULL DEFAULT 0,
  "keep_alive" int4 NOT NULL DEFAULT 0,
  "status" int4 NOT NULL DEFAULT 1,
  "create_time" timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "update_time" timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted" int4 NOT NULL DEFAULT 0
)
;
ALTER TABLE "public"."sys_route" OWNER TO "postgres";
COMMENT ON COLUMN "public"."sys_route"."id" IS '主键';
COMMENT ON COLUMN "public"."sys_route"."parent_id" IS '父级 id，空串表示一级';
COMMENT ON COLUMN "public"."sys_route"."name" IS 'Elegant 路由名，如 home、sys-user';
COMMENT ON COLUMN "public"."sys_route"."path" IS '路径，如 /home';
COMMENT ON COLUMN "public"."sys_route"."component" IS '如 layout.base$view.home';
COMMENT ON COLUMN "public"."sys_route"."status" IS '1 启用 0 禁用';
COMMENT ON TABLE "public"."sys_route" IS '系统菜单/路由';

-- ----------------------------
-- Records of sys_route
-- ----------------------------
BEGIN;
INSERT INTO "public"."sys_route" ("id", "parent_id", "name", "path", "component", "title", "i18n_key", "icon", "order_no", "hide_in_menu", "keep_alive", "status", "create_time", "update_time", "deleted") VALUES ('00000000-0000-0000-0000-000000000010', '', 'home', '/home', 'layout.base$view.home', 'home', 'route.home', 'mdi:monitor-dashboard', 1, 0, 0, 1, '2026-05-15 09:31:47.237353', '2026-05-15 09:31:47.237353', 0);
INSERT INTO "public"."sys_route" ("id", "parent_id", "name", "path", "component", "title", "i18n_key", "icon", "order_no", "hide_in_menu", "keep_alive", "status", "create_time", "update_time", "deleted") VALUES ('00000000-0000-0000-0000-000000000011', '', 'sys-user', '/sys-user', 'layout.base$view.sys-user', '用户管理', 'route.sys-user', 'mdi:account-group', 20, 0, 0, 1, '2026-05-15 09:31:47.237353', '2026-05-15 09:31:47.237353', 0);
INSERT INTO "public"."sys_route" ("id", "parent_id", "name", "path", "component", "title", "i18n_key", "icon", "order_no", "hide_in_menu", "keep_alive", "status", "create_time", "update_time", "deleted") VALUES ('00000000-0000-0000-0000-000000000012', '', 'sys-role', '/sys-role', 'layout.base$view.sys-role', '角色管理', 'route.sys-role', 'mdi:shield-account', 30, 0, 0, 1, '2026-05-15 09:31:47.237353', '2026-05-15 09:31:47.237353', 0);
INSERT INTO "public"."sys_route" ("id", "parent_id", "name", "path", "component", "title", "i18n_key", "icon", "order_no", "hide_in_menu", "keep_alive", "status", "create_time", "update_time", "deleted") VALUES ('00000000-0000-0000-0000-000000000013', '', 'sys-route', '/sys-route', 'layout.base$view.sys-route', '菜单路由', 'route.sys-route', 'mdi:family-tree', 40, 0, 0, 1, '2026-05-15 09:31:47.237353', '2026-05-15 09:31:47.237353', 0);
INSERT INTO "public"."sys_route" ("id", "parent_id", "name", "path", "component", "title", "i18n_key", "icon", "order_no", "hide_in_menu", "keep_alive", "status", "create_time", "update_time", "deleted") VALUES ('00000000-0000-0000-0000-000000000014', '', 'package-mall', '/package-mall', 'layout.base$view.package-mall', '套餐购买', 'route.package-mall', 'mdi:shopping-outline', 8, 0, 0, 1, '2026-05-15 09:31:47.237353', '2026-05-15 09:31:47.237353', 0);
INSERT INTO "public"."sys_route" ("id", "parent_id", "name", "path", "component", "title", "i18n_key", "icon", "order_no", "hide_in_menu", "keep_alive", "status", "create_time", "update_time", "deleted") VALUES ('00000000-0000-0000-0000-000000000015', '', 'sys-package', '/sys-package', 'layout.base$view.sys-package', '套餐配置', 'route.sys-package', 'mdi:package-variant', 45, 0, 0, 1, '2026-05-15 09:31:47.237353', '2026-05-15 09:31:47.237353', 0);
INSERT INTO "public"."sys_route" ("id", "parent_id", "name", "path", "component", "title", "i18n_key", "icon", "order_no", "hide_in_menu", "keep_alive", "status", "create_time", "update_time", "deleted") VALUES ('00000000-0000-0000-0000-000000000016', '', 'sys-announcement', '/sys-announcement', 'layout.base$view.sys-announcement', '公告管理', 'route.sys-announcement', 'mdi:bullhorn-outline', 46, 0, 0, 1, '2026-05-15 09:31:47.237353', '2026-05-15 09:31:47.237353', 0);
INSERT INTO "public"."sys_route" ("id", "parent_id", "name", "path", "component", "title", "i18n_key", "icon", "order_no", "hide_in_menu", "keep_alive", "status", "create_time", "update_time", "deleted") VALUES ('00000000-0000-0000-0000-000000000017', '', 'sys-runtime-config', '/sys-runtime-config', 'layout.base$view.sys-runtime-config', '算法运行时配置', 'route.sys-runtime-config', 'mdi:cog-sync-outline', 47, 0, 0, 1, '2026-05-15 09:31:47.237353', '2026-05-15 09:31:47.237353', 0);
COMMIT;

-- ----------------------------
-- Table structure for sys_user
-- ----------------------------
DROP TABLE IF EXISTS "public"."sys_user";
CREATE TABLE "public"."sys_user" (
  "id" varchar(64) COLLATE "pg_catalog"."default" NOT NULL,
  "username" varchar(64) COLLATE "pg_catalog"."default" NOT NULL,
  "password_hash" varchar(128) COLLATE "pg_catalog"."default" NOT NULL,
  "nickname" varchar(128) COLLATE "pg_catalog"."default" NOT NULL DEFAULT ''::character varying,
  "status" int4 NOT NULL DEFAULT 1,
  "is_admin" int4 NOT NULL DEFAULT 0,
  "create_time" timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "update_time" timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted" int4 NOT NULL DEFAULT 0
)
;
ALTER TABLE "public"."sys_user" OWNER TO "postgres";
COMMENT ON COLUMN "public"."sys_user"."id" IS '主键';
COMMENT ON COLUMN "public"."sys_user"."username" IS '用户名（唯一）';
COMMENT ON COLUMN "public"."sys_user"."password_hash" IS '密码哈希（bcrypt）';
COMMENT ON COLUMN "public"."sys_user"."nickname" IS '昵称';
COMMENT ON COLUMN "public"."sys_user"."status" IS '状态：1启用 0禁用';
COMMENT ON COLUMN "public"."sys_user"."is_admin" IS '是否管理员：1是 0否';
COMMENT ON COLUMN "public"."sys_user"."create_time" IS '创建时间';
COMMENT ON COLUMN "public"."sys_user"."update_time" IS '更新时间';
COMMENT ON COLUMN "public"."sys_user"."deleted" IS '逻辑删除：0否 1是';
COMMENT ON TABLE "public"."sys_user" IS '系统用户';

-- ----------------------------
-- Records of sys_user
-- ----------------------------
BEGIN;
INSERT INTO "public"."sys_user" ("id", "username", "password_hash", "nickname", "status", "is_admin", "create_time", "update_time", "deleted") VALUES ('00000000-0000-0000-0000-000000000001', 'admin', '$2a$10$rs2JIgjJfo4Cm30K4WZ5dOohnTMv4B3/y9oZT.s/4s3g7mb/kkh.i', '管理员', 1, 1, '2026-05-15 09:31:47.296021', '2026-05-15 09:31:47.296021', 0);
INSERT INTO "public"."sys_user" ("id", "username", "password_hash", "nickname", "status", "is_admin", "create_time", "update_time", "deleted") VALUES ('0c3f344d-a617-47f7-b75d-fae33a4a62e2', 'panyang', '$2a$10$9ys0QXKyREmjUzC3jfKRJeDK5PhXK68SubDRwA9nCUd3UyR7t9izC', 'panyang', 1, 0, '0001-01-01 00:00:00', '0001-01-01 00:00:00', 0);
COMMIT;

-- ----------------------------
-- Table structure for sys_user_role
-- ----------------------------
DROP TABLE IF EXISTS "public"."sys_user_role";
CREATE TABLE "public"."sys_user_role" (
  "id" varchar(64) COLLATE "pg_catalog"."default" NOT NULL,
  "user_id" varchar(64) COLLATE "pg_catalog"."default" NOT NULL,
  "role_id" varchar(64) COLLATE "pg_catalog"."default" NOT NULL,
  "create_time" timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "update_time" timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted" int4 NOT NULL DEFAULT 0
)
;
ALTER TABLE "public"."sys_user_role" OWNER TO "postgres";
COMMENT ON COLUMN "public"."sys_user_role"."id" IS '主键';
COMMENT ON COLUMN "public"."sys_user_role"."user_id" IS '用户ID';
COMMENT ON COLUMN "public"."sys_user_role"."role_id" IS '角色ID';
COMMENT ON COLUMN "public"."sys_user_role"."create_time" IS '创建时间';
COMMENT ON COLUMN "public"."sys_user_role"."update_time" IS '更新时间';
COMMENT ON COLUMN "public"."sys_user_role"."deleted" IS '逻辑删除：0否 1是';
COMMENT ON TABLE "public"."sys_user_role" IS '用户角色关联';

-- ----------------------------
-- Records of sys_user_role
-- ----------------------------
BEGIN;
INSERT INTO "public"."sys_user_role" ("id", "user_id", "role_id", "create_time", "update_time", "deleted") VALUES ('f3c89a83-a347-4cbb-86e0-2b52d71602ce', '0c3f344d-a617-47f7-b75d-fae33a4a62e2', '00000000-0000-0000-0000-000000000003', '2026-05-15 09:55:55.714159', '2026-05-15 09:55:55.714159', 0);
INSERT INTO "public"."sys_user_role" ("id", "user_id", "role_id", "create_time", "update_time", "deleted") VALUES ('1fe0282b-2f35-478c-8ac1-99db9cda489f', '00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000002', '2026-05-21 15:05:45.540804', '2026-05-21 15:05:45.540804', 0);
COMMIT;

-- ----------------------------
-- Indexes structure for table biz_announcement
-- ----------------------------
CREATE INDEX "idx_biz_announcement_status" ON "public"."biz_announcement" USING btree (
  "status" "pg_catalog"."int4_ops" ASC NULLS LAST,
  "sort_no" "pg_catalog"."int4_ops" ASC NULLS LAST
) WHERE deleted = 0;

-- ----------------------------
-- Primary Key structure for table biz_announcement
-- ----------------------------
ALTER TABLE "public"."biz_announcement" ADD CONSTRAINT "biz_announcement_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table biz_package
-- ----------------------------
CREATE UNIQUE INDEX "uk_biz_package_code" ON "public"."biz_package" USING btree (
  "code" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
) WHERE deleted = 0;

-- ----------------------------
-- Primary Key structure for table biz_package
-- ----------------------------
ALTER TABLE "public"."biz_package" ADD CONSTRAINT "biz_package_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table biz_usage_log
-- ----------------------------
CREATE INDEX "idx_biz_usage_log_user_time" ON "public"."biz_usage_log" USING btree (
  "user_id" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST,
  "create_time" "pg_catalog"."timestamp_ops" DESC NULLS FIRST
);

-- ----------------------------
-- Primary Key structure for table biz_usage_log
-- ----------------------------
ALTER TABLE "public"."biz_usage_log" ADD CONSTRAINT "biz_usage_log_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Primary Key structure for table biz_user_preference
-- ----------------------------
ALTER TABLE "public"."biz_user_preference" ADD CONSTRAINT "biz_user_preference_pkey" PRIMARY KEY ("user_id");

-- ----------------------------
-- Primary Key structure for table biz_user_quota
-- ----------------------------
ALTER TABLE "public"."biz_user_quota" ADD CONSTRAINT "biz_user_quota_pkey" PRIMARY KEY ("user_id");

-- ----------------------------
-- Indexes structure for table biz_user_voice
-- ----------------------------
CREATE INDEX "idx_biz_user_voice_user" ON "public"."biz_user_voice" USING btree (
  "user_id" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
) WHERE deleted = 0;

-- ----------------------------
-- Primary Key structure for table biz_user_voice
-- ----------------------------
ALTER TABLE "public"."biz_user_voice" ADD CONSTRAINT "biz_user_voice_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table sys_role
-- ----------------------------
CREATE UNIQUE INDEX "uk_sys_role_code" ON "public"."sys_role" USING btree (
  "code" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
) WHERE deleted = 0;

-- ----------------------------
-- Primary Key structure for table sys_role
-- ----------------------------
ALTER TABLE "public"."sys_role" ADD CONSTRAINT "sys_role_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table sys_role_route
-- ----------------------------
CREATE INDEX "idx_sys_role_route_role" ON "public"."sys_role_route" USING btree (
  "role_id" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
) WHERE deleted = 0;
CREATE INDEX "idx_sys_role_route_route" ON "public"."sys_role_route" USING btree (
  "route_id" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
) WHERE deleted = 0;
CREATE UNIQUE INDEX "uk_sys_role_route_pair" ON "public"."sys_role_route" USING btree (
  "role_id" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST,
  "route_id" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
) WHERE deleted = 0;

-- ----------------------------
-- Primary Key structure for table sys_role_route
-- ----------------------------
ALTER TABLE "public"."sys_role_route" ADD CONSTRAINT "sys_role_route_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table sys_route
-- ----------------------------
CREATE UNIQUE INDEX "uk_sys_route_name" ON "public"."sys_route" USING btree (
  "name" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
) WHERE deleted = 0;

-- ----------------------------
-- Primary Key structure for table sys_route
-- ----------------------------
ALTER TABLE "public"."sys_route" ADD CONSTRAINT "sys_route_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table sys_user
-- ----------------------------
CREATE UNIQUE INDEX "uk_sys_user_username" ON "public"."sys_user" USING btree (
  "username" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
) WHERE deleted = 0;

-- ----------------------------
-- Primary Key structure for table sys_user
-- ----------------------------
ALTER TABLE "public"."sys_user" ADD CONSTRAINT "sys_user_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table sys_user_role
-- ----------------------------
CREATE UNIQUE INDEX "uk_sys_user_role_pair" ON "public"."sys_user_role" USING btree (
  "user_id" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST,
  "role_id" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
) WHERE deleted = 0;

-- ----------------------------
-- Primary Key structure for table sys_user_role
-- ----------------------------
ALTER TABLE "public"."sys_user_role" ADD CONSTRAINT "sys_user_role_pkey" PRIMARY KEY ("id");
