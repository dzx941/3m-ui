import React, { useState } from 'react';
import { Button, Card, Form, Input, Typography, message } from 'antd';
import { LockOutlined, UserOutlined } from '@ant-design/icons';
import { Navigate, useLocation, useNavigate } from 'react-router-dom';
import { isAuthenticated, login } from '../api/auth';
import { useI18n } from '../i18n';

const { Title, Paragraph } = Typography;

export default function Login() {
  const { t } = useI18n(); const navigate=useNavigate(); const location=useLocation(); const [loading,setLoading]=useState(false);
  const from=(location.state as {from?:{pathname?:string}}|null)?.from?.pathname||'/dashboard';
  if(isAuthenticated()) return <Navigate to={from} replace/>;
  const submit=async(v:{username:string;password:string})=>{setLoading(true);try{await login(v);message.success(t('login.success'));navigate(from,{replace:true});}catch(e:any){message.error(e.message||t('login.failed'));}finally{setLoading(false);}};
  return <div style={{display:'flex',justifyContent:'center',alignItems:'center',minHeight:'100vh',background:'#f5f7fa'}}>
    <Card style={{width:400,boxShadow:'0 8px 30px rgba(0,0,0,.08)'}}>
      <Title level={3} style={{textAlign:'center',marginBottom:8}}>3m-ui</Title>
      <Paragraph style={{textAlign:'center',marginBottom:24}}>{t('login.subtitle')}</Paragraph>
      <Form layout="vertical" onFinish={submit} requiredMark={false}>
        <Form.Item name="username" label={t('login.username')} rules={[{required:true,message:t('login.requiredUsername')}]}>
          <Input prefix={<UserOutlined/>} autoComplete="username"/>
        </Form.Item>
        <Form.Item name="password" label={t('login.password')} rules={[{required:true,message:t('login.requiredPassword')}]}>
          <Input.Password prefix={<LockOutlined/>} autoComplete="current-password"/>
        </Form.Item>
        <Button type="primary" htmlType="submit" block loading={loading}>{t('login.submit')}</Button>
      </Form>
    </Card>
  </div>;
}
