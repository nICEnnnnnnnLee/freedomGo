package utils

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"os"
	"time"
)

type TlsCert struct {
	CertPath         string
	KeyPath          string
	certificate      *tls.Certificate
	certLastModified *time.Time
	certExpireTime   *time.Time
	lastAttemp       *time.Time
	AttempDuration   time.Duration
}

func (t *TlsCert) LoadCert() {
	now := time.Now()
	t.lastAttemp = &now
	_, err := t.loadCert(nil)
	if err != nil {
		panic(err)
	}
}

func (t *TlsCert) loadCert(certLastModified *time.Time) (*tls.Certificate, error) {
	// x509.ParseCertificate()
	cert, err := tls.LoadX509KeyPair(t.CertPath, t.KeyPath)
	if err != nil {
		return nil, err
	}
	x509Cert, _ := x509.ParseCertificate(cert.Certificate[0])
	log.Println("证书过期时间: ", x509Cert.NotAfter)
	if certLastModified == nil {
		certFileInfo, err := os.Stat(t.CertPath)
		if err != nil {
			return nil, err
		}
		modTime := certFileInfo.ModTime()
		t.certLastModified = &modTime
	} else {
		t.certLastModified = certLastModified
	}

	t.certExpireTime = &x509Cert.NotAfter
	t.certificate = &cert
	return t.certificate, nil
}

func (t *TlsCert) CheckCert(now *time.Time) {
	t.lastAttemp = now
	// // 检查证书有没有过期
	// if !now.After(*t.certExpireTime) {
	// 	// log.Println("证书未过期, 返回旧的证书")
	// 	return
	// }
	// log.Println("证书已过期")

	// 检查文件存在
	certFileInfo, err := os.Stat(t.CertPath)
	if err != nil {
		log.Println("证书不存在, 返回旧的证书:", err)
		return
	}
	// 检查文件有过修改
	certLastModified := certFileInfo.ModTime()
	// if !certLastModified.After(*t.certLastModified) {
	if certLastModified.Compare(*t.certLastModified) == 0 {
		log.Println("证书最后修改时间不合适, 返回旧的证书")
		return
	}
	// 重新加载Cert
	log.Println("尝试加载新证书:", err)
	t.loadCert(&certLastModified)
}

func (t *TlsCert) GetCertFunc() func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	t.LoadCert()
	return func(_ *tls.ClientHelloInfo) (*tls.Certificate, error) {
		// 不论 ClientHello ServerName是什么，都返回该证书
		// 初始化证书  必须在这之前显式调用 t.LoadCert()
		now := time.Now()
		if t.lastAttemp.Add(t.AttempDuration).Before(now) && now.After(*t.certExpireTime) {
			log.Println("证书已过期, 检查是否更新证书")
			t.CheckCert(&now)
		}
		return t.certificate, nil
	}
}
