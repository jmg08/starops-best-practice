package com.alibaba.cloud.cms.samples.client;

import io.github.cdimascio.dotenv.Dotenv;
import java.util.ArrayList;
import java.util.List;

/**
 * 应用配置
 * Application configuration
 */
public class Config {
    private String workspace;
    private String endpoint;
    private String region;
    private String accessKeyId;
    private String accessKeySecret;
    private String employeeName;

    public Config() {}

    /**
     * 从环境变量加载配置
     * Load configuration from environment variables
     */
    public static Config loadFromEnv() throws SDKException {
        Dotenv dotenv = Dotenv.configure()
                .ignoreIfMissing()
                .load();

        Config cfg = new Config();
        cfg.workspace = getEnvValue(dotenv, "VIBEOPS_WORKSPACE");
        cfg.endpoint = getEnvValue(dotenv, "VIBEOPS_ENDPOINT");
        cfg.region = getEnvValue(dotenv, "VIBEOPS_REGION");
        cfg.accessKeyId = getEnvValue(dotenv, "ALIBABA_CLOUD_ACCESS_KEY_ID");
        cfg.accessKeySecret = getEnvValue(dotenv, "ALIBABA_CLOUD_ACCESS_KEY_SECRET");
        cfg.employeeName = getEnvValue(dotenv, "VIBEOPS_EMPLOYEE_NAME");

        // Validate required fields
        List<String> missingVars = new ArrayList<>();
        if (cfg.workspace == null || cfg.workspace.isEmpty()) {
            missingVars.add("VIBEOPS_WORKSPACE");
        }
        if (cfg.endpoint == null || cfg.endpoint.isEmpty()) {
            missingVars.add("VIBEOPS_ENDPOINT");
        }
        if (cfg.accessKeyId == null || cfg.accessKeyId.isEmpty()) {
            missingVars.add("ALIBABA_CLOUD_ACCESS_KEY_ID");
        }
        if (cfg.accessKeySecret == null || cfg.accessKeySecret.isEmpty()) {
            missingVars.add("ALIBABA_CLOUD_ACCESS_KEY_SECRET");
        }

        if (!missingVars.isEmpty()) {
            throw SDKException.configMissing(missingVars);
        }

        // Set defaults
        if (cfg.employeeName == null || cfg.employeeName.isEmpty()) {
            cfg.employeeName = "default";
        }
        if (cfg.region == null || cfg.region.isEmpty()) {
            cfg.region = "cn-hangzhou";
        }

        return cfg;
    }

    private static String getEnvValue(Dotenv dotenv, String key) {
        String value = dotenv.get(key);
        if (value == null || value.isEmpty()) {
            value = System.getenv(key);
        }
        return value;
    }

    // Getters and Setters
    public String getWorkspace() {
        return workspace;
    }

    public void setWorkspace(String workspace) {
        this.workspace = workspace;
    }

    public String getEndpoint() {
        return endpoint;
    }

    public void setEndpoint(String endpoint) {
        this.endpoint = endpoint;
    }

    public String getRegion() {
        return region;
    }

    public void setRegion(String region) {
        this.region = region;
    }

    public String getAccessKeyId() {
        return accessKeyId;
    }

    public void setAccessKeyId(String accessKeyId) {
        this.accessKeyId = accessKeyId;
    }

    public String getAccessKeySecret() {
        return accessKeySecret;
    }

    public void setAccessKeySecret(String accessKeySecret) {
        this.accessKeySecret = accessKeySecret;
    }

    public String getEmployeeName() {
        return employeeName;
    }

    public void setEmployeeName(String employeeName) {
        this.employeeName = employeeName;
    }
}
