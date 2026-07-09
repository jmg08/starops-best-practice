/**
 * 默认凭据链支持 - 与 Go 实现保持一致
 * 凭据链优先级：环境变量 > OIDC > CLI配置文件 > 配置文件(~/.alibabacloud/credentials) > IAM角色
 */
import Credential from '@alicloud/credentials';

export async function loadCredentialsFromChain(): Promise<{ accessKeyId: string; accessKeySecret: string }> {
  // undefined config triggers default credential chain
  const cred = new Credential();
  const credential = await cred.getCredential();
  const accessKeyId = credential.accessKeyId;
  const accessKeySecret = credential.accessKeySecret;
  if (!accessKeyId || !accessKeySecret) {
    throw new Error('凭据链返回的凭证为空');
  }
  return { accessKeyId, accessKeySecret };
}
