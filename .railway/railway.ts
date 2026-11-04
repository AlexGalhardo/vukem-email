import { defineRailway, postgres, preserve, project, service } from 'railway/iac';
export default defineRailway(() => {
 const db = postgres('postgres');
 const env = {
  APP_ENV:'production', APP_DEBUG:'false', APP_KEY:preserve(), APP_URL:preserve(),
  DATABASE_URL:db.env.DATABASE_URL,
  SOCIAL_GOOGLE_CLIENT_ID:preserve(), SOCIAL_GOOGLE_CLIENT_SECRET:preserve(),
  RESEND_API_KEY:preserve(), MAIL_MAILER:'resend', MAIL_FROM:preserve(),
  CACHE_STORE:'database', SESSION_DRIVER:'cookie',
  QUEUE_CONNECTION:'database', QUEUE_CONCURRENCY:'1', QUEUE_IN_SERVE:'false',
  TRUSTED_PROXIES:'100.64.0.0/24', CLIENT_IP_HEADER:'X-Real-IP', TRUST_FORWARDED_PROTO:'true',
 };
 const web = service('vukem-email', {healthcheck:'/up', preDeploy:'/app/app migrate', env});
 const worker = service('vukem-email-worker', {start:'/app/app queue:work', env});
 return project('vukem-email', {resources:[db,web,worker]});
});
