import React, { useState, useEffect } from 'react';
import { Typography, Row, Col, Card, Button, Space, Tag, message, Descriptions } from 'antd';
import {
  SaveOutlined,
  CheckCircleOutlined,
  DownloadOutlined,
  ReloadOutlined,
  SettingOutlined,
  ExclamationCircleOutlined,
} from '@ant-design/icons';
import Editor from '@monaco-editor/react';

const { Title, Paragraph } = Typography;

const API_BASE = 'http://localhost:8080/api/v1';

const Config: React.FC = () => {
  const [yamlCode, setYamlCode] = useState<string>('');
  const [loading, setLoading] = useState<boolean>(false);
  const [validating, setValidating] = useState<boolean>(false);
  const [isValid, setIsValid] = useState<boolean | null>(null);
  const [validationError, setValidationError] = useState<string>('');

  const fetchConfig = async () => {
    setLoading(true);
    try {
      const res = await fetch(`${API_BASE}/config`);
      if (res.ok) {
        const data = await res.json();
        setYamlCode(data.config || '');
        // Validate initially
        runValidate(data.config || '');
      } else {
        message.error('Failed to load current configuration.');
      }
    } catch (err) {
      message.error('Backend service unreachable.');
    } finally {
      setLoading(false);
    }
  };

  const runValidate = async (codeToValidate: string) => {
    setValidating(true);
    try {
      const res = await fetch(`${API_BASE}/config/validate`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ config: codeToValidate }),
      });
      if (res.ok) {
        const data = await res.json();
        setIsValid(data.valid);
        setValidationError(data.valid ? '' : data.error || 'Syntax checking failed');
      } else {
        message.error('Validation API request failed.');
      }
    } catch (err) {
      message.error('Unable to connect to validator API.');
    } finally {
      setValidating(false);
    }
  };

  useEffect(() => {
    fetchConfig();
  }, []);

  const handleSaveGenerate = async () => {
    setLoading(true);
    try {
      const res = await fetch(`${API_BASE}/config/generate`, {
        method: 'POST',
      });
      if (res.ok) {
        message.success('Configuration saved and generated successfully!');
        await fetchConfig();
      } else {
        message.error('Failed to write configuration file.');
      }
    } catch (err) {
      message.error('Network connection error.');
    } finally {
      setLoading(false);
    }
  };

  const handleDownload = () => {
    window.open(`${API_BASE}/config/download`, '_blank');
    message.success('Downloading config.yaml...');
  };

  const handleHotReload = async () => {
    setLoading(true);
    try {
      const res = await fetch(`${API_BASE}/mihomo/restart`, {
        method: 'POST',
      });
      if (res.ok) {
        message.success('Mihomo service reloaded configuration!');
      } else {
        message.error('Failed to trigger Mihomo reload.');
      }
    } catch (err) {
      message.error('Network connection error.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <div>
          <Title level={2} style={{ margin: 0 }}>Mihomo Config Engine</Title>
          <Paragraph style={{ margin: 0 }}>
            Inspect, modify, validate, and hot-reload dynamic Mihomo VPS configuration parameters.
          </Paragraph>
        </div>
      </div>

      <Row gutter={[24, 24]}>
        {/* Left Panel: Monaco Code Editor */}
        <Col xs={24} lg={16}>
          <Card
            title={
              <Space>
                <SettingOutlined />
                <span>config.yaml (Final Merged View)</span>
              </Space>
            }
            bordered={false}
            bodyStyle={{ padding: 0 }}
          >
            <div style={{ border: '1px solid #d9d9d9', borderRadius: '0 0 8px 8px', overflow: 'hidden' }}>
              <Editor
                height="60vh"
                language="yaml"
                theme="vs-light"
                value={yamlCode}
                onChange={(val) => {
                  const newVal = val || '';
                  setYamlCode(newVal);
                  runValidate(newVal);
                }}
                options={{
                  minimap: { enabled: false },
                  fontSize: 14,
                  wordWrap: 'on',
                  scrollBeyondLastLine: false,
                  automaticLayout: true,
                }}
              />
            </div>
          </Card>
        </Col>

        {/* Right Panel: Control panel Actions & Validation metrics */}
        <Col xs={24} lg={8}>
          <Space direction="vertical" size="large" style={{ width: '100%' }}>
            <Card title="Config Control Center" bordered={false}>
              <Space direction="vertical" style={{ width: '100%' }} size="middle">
                <Button
                  type="primary"
                  icon={<SaveOutlined />}
                  block
                  loading={loading}
                  onClick={handleSaveGenerate}
                >
                  Save & Write config.yaml
                </Button>
                <Button
                  icon={<CheckCircleOutlined />}
                  block
                  loading={validating}
                  onClick={() => runValidate(yamlCode)}
                >
                  Validate Config
                </Button>
                <Button
                  icon={<DownloadOutlined />}
                  block
                  onClick={handleDownload}
                >
                  Download config.yaml
                </Button>
                <Button
                  icon={<ReloadOutlined />}
                  danger
                  block
                  loading={loading}
                  onClick={handleHotReload}
                >
                  Trigger Hot Reload
                </Button>
              </Space>
            </Card>

            <Card title="Validation Engine Status" bordered={false}>
              <Descriptions column={1} bordered size="small">
                <Descriptions.Item label="Engine Verification">
                  {isValid === null ? (
                    <Tag color="default">Checking...</Tag>
                  ) : isValid ? (
                    <Tag color="success" icon={<CheckCircleOutlined />}>
                      Valid YAML Schema
                    </Tag>
                  ) : (
                    <Tag color="error" icon={<ExclamationCircleOutlined />}>
                      Invalid Schema
                    </Tag>
                  )}
                </Descriptions.Item>
              </Descriptions>

              {isValid === false && (
                <div style={{ marginTop: 16, padding: 12, backgroundColor: '#fff2f0', border: '1px solid #ffccc7', borderRadius: 4, color: '#ff4d4f', fontFamily: 'monospace', fontSize: '13px' }}>
                  <strong>Error details:</strong>
                  <div style={{ marginTop: 4 }}>{validationError}</div>
                </div>
              )}
            </Card>
          </Space>
        </Col>
      </Row>
    </div>
  );
};

export default Config;
