<template>
  <div class="settings-page">
    <el-tabs v-model="activeTab" type="border-card">
      <el-tab-pane label="界面配置" name="ui">
        <el-card class="config-card">
          <template #header>
            <span>基本信息</span>
          </template>
          <el-form :model="uiForm" label-width="120px" class="config-form">
            <el-row :gutter="24">
              <el-col :span="12">
                <el-form-item label="浏览器标题">
                  <el-input v-model="uiForm.browserTitle" placeholder="能耗管理中心" />
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="登录页标题">
                  <el-input v-model="uiForm.loginTitle" placeholder="能耗管理系统" />
                </el-form-item>
              </el-col>
            </el-row>
            <el-form-item label="Logo URL">
              <el-input v-model="uiForm.logo" placeholder="可选，留空使用默认Logo" />
            </el-form-item>
          </el-form>
        </el-card>

        <el-card class="config-card">
          <template #header>
            <span>页脚配置</span>
          </template>
          <el-form :model="uiForm" label-width="120px" class="config-form">
            <el-row :gutter="24">
              <el-col :span="12">
                <el-form-item label="公司简称">
                  <el-input v-model="uiForm.footerShortName" placeholder="如：xxGroup" />
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="公司名称">
                  <el-input v-model="uiForm.footerCompany" placeholder="如：xx科技有限公司" />
                </el-form-item>
              </el-col>
            </el-row>
            <el-form-item label="ICP备案号">
              <el-input v-model="uiForm.footerICP" placeholder="如：浙ICP备xxxxxxxx号" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="saveUI" :loading="savingUI">保存配置</el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="HTTPS证书" name="https">
        <el-card class="config-card">
          <template #header>
            <div class="card-header">
              <span>SSL/TLS证书配置</span>
              <el-tag :type="httpsForm.enabled ? 'success' : 'info'" size="small">
                {{ httpsForm.enabled ? 'HTTPS已启用' : 'HTTPS未启用' }}
              </el-tag>
            </div>
          </template>
          
          <!-- 证书状态 -->
          <div class="cert-status" v-if="httpsForm.certExists && httpsForm.keyExists">
            <el-alert type="success" :closable="false" show-icon>
              <template #title>证书已上传</template>
              <template #default>
                <div class="cert-info">
                  <p><strong>域名：</strong>{{ httpsForm.domain || '未知' }}</p>
                  <p><strong>过期时间：</strong>{{ formatExpiry(httpsForm.certExpiry) }}</p>
                  <p><strong>证书主题：</strong>{{ httpsForm.certSubject || '未知' }}</p>
                </div>
              </template>
            </el-alert>
          </div>
          <div class="cert-status" v-else>
            <el-alert type="warning" :closable="false" show-icon>
              <template #title>未上传证书</template>
              <template #default>请上传SSL证书和私钥文件以启用HTTPS</template>
            </el-alert>
          </div>

          <el-divider />

          <!-- HTTPS配置 -->
          <el-form :model="httpsForm" label-width="120px" class="config-form">
            <el-row :gutter="24">
              <el-col :span="12">
                <el-form-item label="启用HTTPS">
                  <el-switch v-model="httpsForm.enabled" :disabled="!httpsForm.certExists || !httpsForm.keyExists" />
                  <div class="form-tip" v-if="!httpsForm.certExists">请先上传证书</div>
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="HTTPS端口">
                  <el-input v-model="httpsForm.port" placeholder="8443" />
                </el-form-item>
              </el-col>
            </el-row>
            <el-form-item>
              <el-button type="primary" @click="saveHttps" :loading="savingHttps">保存配置</el-button>
            </el-form-item>
          </el-form>
        </el-card>

        <el-card class="config-card">
          <template #header>
            <span>上传证书</span>
          </template>
          <el-form label-width="120px" class="config-form">
            <el-row :gutter="24">
              <el-col :span="12">
                <el-form-item label="证书文件">
                  <div class="upload-wrapper">
                    <el-upload
                      ref="certUploadRef"
                      :auto-upload="false"
                      :limit="1"
                      :on-change="(file: any) => certFile = file.raw"
                      :on-remove="() => certFile = null"
                      accept=".crt,.pem,.cer"
                    >
                      <template #trigger>
                        <el-button type="primary">选择证书</el-button>
                      </template>
                    </el-upload>
                    <span class="upload-format">.crt, .pem, .cer</span>
                  </div>
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="私钥文件">
                  <div class="upload-wrapper">
                    <el-upload
                      ref="keyUploadRef"
                      :auto-upload="false"
                      :limit="1"
                      :on-change="(file: any) => keyFile = file.raw"
                      :on-remove="() => keyFile = null"
                      accept=".key,.pem"
                    >
                      <template #trigger>
                        <el-button type="primary">选择私钥</el-button>
                      </template>
                    </el-upload>
                    <span class="upload-format">.key, .pem</span>
                  </div>
                </el-form-item>
              </el-col>
            </el-row>
            <el-form-item>
              <el-button type="success" @click="uploadCert" :loading="uploading" :disabled="!certFile || !keyFile">
                上传证书
              </el-button>
              <el-button type="danger" @click="deleteCert" v-if="httpsForm.certExists" :loading="deleting">
                删除证书
              </el-button>
            </el-form-item>
          </el-form>

          <el-divider />

          <div class="help-content">
            <h4>配置说明</h4>
            <ol>
              <li>准备SSL证书文件（.crt/.pem）和私钥文件（.key/.pem）</li>
              <li>上传证书和私钥文件</li>
              <li>配置HTTPS端口（默认8443）</li>
              <li>启用HTTPS并保存配置</li>
              <li>重启服务后HTTPS生效</li>
            </ol>
          </div>
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="文档" name="docs">
        <div class="docs-header">
          <p class="docs-desc">可以下载本系统的所有相关文档，包括API文档、系统文档、帮助文档</p>
        </div>
        <div class="docs-grid">
          <el-card v-for="doc in docList" :key="doc.id" class="doc-card" shadow="hover">
            <div class="doc-card-body">
              <div class="doc-icon">
                <el-icon :size="36" :style="{ color: 'var(--color-primary)' }"><component :is="doc.icon" /></el-icon>
              </div>
              <div class="doc-info">
                <h3 class="doc-name">{{ doc.name }}</h3>
                <p class="doc-description">{{ doc.description }}</p>
                <p class="doc-filename">{{ doc.filename }}</p>
              </div>
              <div class="doc-action">
                <el-button type="primary" @click="previewDoc(doc.id)">
                  <el-icon><View /></el-icon>预览
                </el-button>
                <el-button @click="downloadDoc(doc.id, doc.filename)">
                  <el-icon><Download /></el-icon>下载
                </el-button>
              </div>
            </div>
          </el-card>
        </div>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { Download, View, Setting, Document, Connection } from "@element-plus/icons-vue";
import { settingsApi, api } from "../../api";
import axios from "axios";

const activeTab = ref("ui");
const savingUI = ref(false);
const savingHttps = ref(false);
const uploading = ref(false);
const deleting = ref(false);

const uiForm = reactive({
  browserTitle: "",
  loginTitle: "",
  logo: "",
  footerShortName: "",
  footerCompany: "",
  footerICP: ""
});

const httpsForm = reactive({
  enabled: false,
  port: "8443",
  domain: "",
  certExpiry: "",
  certSubject: "",
  certExists: false,
  keyExists: false
});

const certFile = ref<File | null>(null);
const keyFile = ref<File | null>(null);
const certUploadRef = ref();
const keyUploadRef = ref();

const loadUI = async () => {
  try {
    const res = await settingsApi.getUI();
    if (res.data.success) {
      Object.assign(uiForm, res.data.data);
    }
  } catch (e) {}
};

const saveUI = async () => {
  savingUI.value = true;
  try {
    await settingsApi.updateUI(uiForm);
    ElMessage.success("保存成功");
    if (uiForm.browserTitle) {
      document.title = uiForm.browserTitle;
    }
  } finally {
    savingUI.value = false;
  }
};

const loadHttps = async () => {
  try {
    const res = await settingsApi.getHttps();
    if (res.data.success && res.data.data) {
      Object.assign(httpsForm, res.data.data);
    }
  } catch (e) {}
};

const saveHttps = async () => {
  savingHttps.value = true;
  try {
    await settingsApi.updateHttps({
      enabled: httpsForm.enabled,
      port: httpsForm.port
    });
    ElMessage.success("保存成功，重启服务后生效");
  } finally {
    savingHttps.value = false;
  }
};

const uploadCert = async () => {
  if (!certFile.value || !keyFile.value) {
    ElMessage.warning("请选择证书文件和私钥文件");
    return;
  }

  uploading.value = true;
  try {
    const formData = new FormData();
    formData.append("cert", certFile.value);
    formData.append("key", keyFile.value);

    const res = await settingsApi.uploadCert(formData);
    if (res.data.success) {
      ElMessage.success("证书上传成功");
      // 清除上传组件的文件列表
      certUploadRef.value?.clearFiles();
      keyUploadRef.value?.clearFiles();
      // 重置文件状态
      certFile.value = null;
      keyFile.value = null;
      // 重新加载证书信息
      await loadHttps();
    } else {
      ElMessage.error(res.data.message || "证书上传失败");
    }
  } catch (e: any) {
    const errMsg = e?.response?.data?.message || "证书上传失败";
    ElMessage.error(errMsg);
  } finally {
    uploading.value = false;
  }
};

const deleteCert = async () => {
  try {
    await ElMessageBox.confirm("确定要删除SSL证书吗？删除后HTTPS将无法使用。", "确认删除", {
      type: "warning"
    });
    
    deleting.value = true;
    const res = await settingsApi.deleteCert();
    if (res.data.success) {
      ElMessage.success("证书已删除");
      await loadHttps();
    }
  } catch (e) {
    // 取消删除
  } finally {
    deleting.value = false;
  }
};

const formatExpiry = (expiry: string) => {
  if (!expiry) return "未知";
  try {
    const date = new Date(expiry);
    const now = new Date();
    const diff = date.getTime() - now.getTime();
    const days = Math.floor(diff / (1000 * 60 * 60 * 24));
    
    const formatted = date.toLocaleDateString("zh-CN");
    
    if (days < 0) {
      return `${formatted} (已过期)`;
    } else if (days < 30) {
      return `${formatted} (剩余 ${days} 天)`;
    }
    return formatted;
  } catch {
    return expiry;
  }
};

// ========== 文档管理 ==========
const docList = ref<any[]>([]);

const loadDocs = async () => {
  try {
    const res = await api.get('/docs');
    if (res.data?.success) {
      docList.value = res.data.data || [];
    }
  } catch {
    docList.value = [
      { id: 'technical', name: '技术架构文档', description: '系统功能说明、代码结构、数据库设计和服务依赖需求', filename: '技术架构文档.pdf', icon: 'Setting' },
      { id: 'manual', name: '系统使用手册', description: '所有功能模块的操作指南和使用示例说明', filename: '系统使用手册.pdf', icon: 'Document' },
      { id: 'api', name: 'API接口文档', description: '完整的REST API调用文档，含示例和错误码', filename: 'API接口文档.pdf', icon: 'Connection' },
    ];
  }
};

const downloadDoc = async (docId: string, filename: string) => {
  try {
    const res = await fetch(`/api/docs/${docId}`, {
      headers: { Authorization: `Bearer ${localStorage.getItem('token')}` }
    });
    if (!res.ok) throw new Error('下载失败');
    const blob = await res.blob();
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    a.click();
    URL.revokeObjectURL(url);
    ElMessage.success('下载成功 ' + filename);
  } catch {
    ElMessage.error('文档下载失败，请确认已登录');
  }
};

const previewDoc = async (docId: string) => {
  try {
    const res = await fetch(`/api/docs/${docId}?mode=preview`, {
      headers: { Authorization: `Bearer ${localStorage.getItem('token')}` }
    });
    if (!res.ok) throw new Error('预览失败');
    const blob = await res.blob();
    const url = URL.createObjectURL(blob);
    window.open(url, '_blank');
  } catch {
    ElMessage.error('文档预览失败，请确认已登录');
  }
};

onMounted(() => {
  loadUI();
  loadHttps();
  loadDocs();
});
</script>

<style scoped>
.settings-page {
  max-width: 900px;
}

.config-card {
  margin-bottom: var(--spacing-xl);
}

.config-card :deep(.el-card__header) {
  padding: 14px 20px;
  background: var(--color-fill-secondary);
  font-weight: 500;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.config-form {
  padding: 10px 0;
}

.form-tip {
  font-size: 12px;
  color: var(--color-text-tertiary);
  margin-top: 4px;
}

.cert-status {
  margin-bottom: 16px;
}

.cert-info p {
  margin: 4px 0;
  font-size: 13px;
}

.upload-tip {
  font-size: 12px;
  color: var(--color-text-tertiary);
  margin-top: 4px;
}

.help-content {
  background: var(--color-fill-secondary);
  border-radius: var(--radius-lg);
  padding: 16px;
}

.help-content h4 {
  margin: 0 0 12px 0;
  font-size: 14px;
  color: var(--color-text-primary);
}

.help-content ol {
  margin: 0;
  padding-left: 20px;
  color: var(--color-text-secondary);
  font-size: 13px;
  line-height: 1.8;
}

/* 文档Tab样式 */
.docs-header {
  margin-bottom: var(--spacing-2xl);
}

.docs-desc {
  font-size: 14px;
  color: var(--color-text-tertiary);
  margin: 0;
}

.docs-grid {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.doc-card {
  border-radius: var(--radius-lg);
  transition: all 0.3s;
}

.doc-card:hover {
  border-color: var(--color-primary);
}

.doc-card :deep(.el-card__body) {
  padding: 20px 24px;
}

.doc-card-body {
  display: flex;
  align-items: center;
  gap: 20px;
}

.doc-icon {
  flex-shrink: 0;
  width: 60px;
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color-primary-bg);
  border-radius: var(--radius-xl);
}

.doc-info {
  flex: 1;
  min-width: 0;
}

.doc-name {
  margin: 0 0 6px 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--color-text-primary);
}

.doc-description {
  margin: 0 0 4px 0;
  font-size: 13px;
  color: var(--color-text-secondary);
  line-height: 1.5;
}

.doc-filename {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-tertiary);
}

.doc-action {
  flex-shrink: 0;
}

/* 修复输入框在深色背景上的可见性 */
.settings-page .el-input__wrapper {
  background-color: #ffffff !important;
  box-shadow: 0 0 0 1px #dcdfe6 inset !important;
}

.settings-page .el-input__wrapper:hover {
  box-shadow: 0 0 0 1px #c0c4cc inset !important;
}

.settings-page .el-input__wrapper.is-focus {
  box-shadow: 0 0 0 1px #667eea inset !important;
}

.settings-page .el-input__inner {
  color: #303133 !important;
  background-color: transparent !important;
}

.settings-page .el-textarea__inner {
  background-color: #ffffff !important;
  color: #303133 !important;
  border: 1px solid #dcdfe6 !important;
}

.settings-page .el-textarea__inner:hover {
  border-color: #c0c4cc !important;
}

.settings-page .el-textarea__inner:focus {
  border-color: #667eea !important;
}

/* 修复 Select 下拉框 */
.settings-page .el-select .el-input__wrapper {
  background-color: #ffffff !important;
}

/* 修复 InputNumber */
.settings-page .el-input-number {
  --el-input-bg-color: #ffffff !important;
}

.settings-page .el-input-number .el-input__wrapper {
  background-color: #ffffff !important;
}

/* 修复上传组件 */
.settings-page .el-upload .el-input__wrapper {
  background-color: #ffffff !important;
  box-shadow: 0 0 0 1px #dcdfe6 inset !important;
}

.settings-page .el-upload .el-input__wrapper:hover {
  box-shadow: 0 0 0 1px #c0c4cc inset !important;
}

.settings-page .el-upload .el-input__wrapper.is-focus {
  box-shadow: 0 0 0 1px #667eea inset !important;
}

.settings-page .el-upload .el-input__inner {
  color: #303133 !important;
  background-color: transparent !important;
}

/* 修复上传按钮 */
.settings-page .el-upload__tip {
  color: #868e96 !important;
  font-size: 12px !important;
}

/* 修复上传触发按钮 */
.settings-page .el-upload .el-button--primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%) !important;
  border: none !important;
  border-radius: 6px !important;
}

.settings-page .el-upload .el-button--primary:hover {
  background: linear-gradient(135deg, #5a6fd6 0%, #6a4190 100%) !important;
}

/* 上传包装器样式 */
.upload-wrapper {
  display: flex;
  align-items: center;
  gap: 12px;
}

.upload-format {
  color: #868e96;
  font-size: 12px;
}

/* 操作按钮样式 */
.config-card .el-button--success {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%) !important;
  border: none !important;
  border-radius: 6px !important;
}

.config-card .el-button--success:hover {
  background: linear-gradient(135deg, #5a6fd6 0%, #6a4190 100%) !important;
}

.config-card .el-button--danger {
  background: #fa5252 !important;
  border: none !important;
  border-radius: 6px !important;
}

.config-card .el-button--danger:hover {
  background: #e03131 !important;
}
</style>
