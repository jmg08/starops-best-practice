/**
 * Configuration for STAROps SDK
 * STAROps SDK 配置
 */

import { config as dotenvConfig } from 'dotenv';
import { SDKException } from './errors.js';

/** 应用配置 / Application configuration */
export interface Config {
  workspace: string;
  endpoint: string;
  accessKeyId: string;
  accessKeySecret: string;
  region: string;
  employeeName: string;
}

/** 从环境变量加载配置 / Load configuration from environment variables */
export function loadConfigFromEnv(): Config {
  dotenvConfig();

  const workspace = process.env.VIBEOPS_WORKSPACE || '';
  const endpoint = process.env.VIBEOPS_ENDPOINT || '';
  const region = process.env.VIBEOPS_REGION || 'cn-hangzhou';
  const accessKeyId = process.env.ALIBABA_CLOUD_ACCESS_KEY_ID || '';
  const accessKeySecret = process.env.ALIBABA_CLOUD_ACCESS_KEY_SECRET || '';
  const employeeName = process.env.VIBEOPS_EMPLOYEE_NAME || 'default';

  // Validate required fields
  const missingVars: string[] = [];
  if (!endpoint) missingVars.push('VIBEOPS_ENDPOINT');
  if (!accessKeyId) missingVars.push('ALIBABA_CLOUD_ACCESS_KEY_ID');
  if (!accessKeySecret) missingVars.push('ALIBABA_CLOUD_ACCESS_KEY_SECRET');

  if (missingVars.length > 0) {
    throw SDKException.configMissing(missingVars);
  }

  return {
    workspace,
    endpoint,
    accessKeyId,
    accessKeySecret,
    region: region || 'cn-hangzhou',
    employeeName: employeeName || 'default',
  };
}
