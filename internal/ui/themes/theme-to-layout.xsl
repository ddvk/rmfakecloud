<?xml version="1.0" encoding="UTF-8"?>
<xsl:stylesheet version="1.0" xmlns:xsl="http://www.w3.org/1999/XSL/Transform">
  <xsl:output method="xml" encoding="UTF-8" indent="yes"/>
  <xsl:template match="/theme">
    <layout id="{@id}" name="{@name}">
      <icons
        brand="{icons/@brand}"
        documents="{icons/@documents}"
        integrations="{icons/@integrations}"
        connect="{icons/@connect}"
        screenshare="{icons/@screenshare}"
        templates="{icons/@templates}"
        admin="{icons/@admin}"
        profile="{icons/@profile}"
      />
      <chrome
        style="{layout/chrome/@style}"
        folder-color="{layout/chrome/@folder-color}"
        tab-color="{layout/chrome/@tab-color}"
        outline-color="{layout/chrome/@outline-color}"
        drawer-color="{layout/chrome/@drawer-color}"
        label-color="{layout/chrome/@label-color}"
      />
      <nav brand-position="{layout/nav/@brand-position}" user-menu="{layout/nav/@user-menu}">
        <xsl:for-each select="layout/nav/item">
          <item
            id="{@id}"
            visible="{@visible}"
            admin-only="{@admin-only}"
            icon="{/theme/icons/@*[local-name()=current()/@id]}"
          >
            <xsl:attribute name="href">
              <xsl:choose>
                <xsl:when test="@id = 'templates'">/admin/templates</xsl:when>
                <xsl:when test="@id = 'admin'">/admin</xsl:when>
                <xsl:when test="@id = 'help'">/help</xsl:when>
                <xsl:when test="@id = 'profile'">/profile</xsl:when>
                <xsl:when test="@id = 'screenshare'">/screenshare</xsl:when>
                <xsl:otherwise>/<xsl:value-of select="@id"/></xsl:otherwise>
              </xsl:choose>
            </xsl:attribute>
          </item>
        </xsl:for-each>
      </nav>
      <login
        primary-button="{layout/login/@primary-button}"
        passkey-button="{layout/login/@passkey-button}"
        show-brand="{layout/login/@show-brand}"
      />
      <connect
        gauge-type="{layout/connect/@gauge-type}"
        track-color="{layout/connect/@track-color}"
        fill-color="{layout/connect/@fill-color}"
        text-color="{layout/connect/@text-color}"
        prompt-location="{layout/connect/@prompt-location}"
        prompt-style="{layout/connect/@prompt-style}"
        prompt-bg="{layout/connect/@prompt-bg}"
        prompt-fg="{layout/connect/@prompt-fg}"
        prompt-border="{layout/connect/@prompt-border}"
        font-family="{layout/connect/@font-family}"
        font-size="{layout/connect/@font-size}"
        code-color="{layout/connect/@code-color}"
      />
    </layout>
  </xsl:template>
</xsl:stylesheet>
