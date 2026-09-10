<?xml version="1.0" encoding="UTF-8"?>
<xsl:stylesheet version="1.0" xmlns:xsl="http://www.w3.org/1999/XSL/Transform">
  <!--
    Transforms a <page> XML document into W3C-valid HTML5.
    XSLT 1.0 only. Go may also prepend <!DOCTYPE html> if using omit-xml-declaration.
  -->
  <xsl:output
    method="html"
    encoding="UTF-8"
    indent="yes"
    omit-xml-declaration="yes"
    doctype-system="about:legacy-compat"
  />

  <xsl:variable name="isAdmin" select="/page/user/@admin = 'true'"/>
  <xsl:variable name="chrome">
    <xsl:choose>
      <xsl:when test="/page/@chrome-style = 'remarkable' or /page/@chrome-style = 'googledocs' or /page/@chrome-style = 'icloud' or /page/@chrome-style = 'os'">
        <xsl:value-of select="/page/@chrome-style"/>
      </xsl:when>
      <xsl:otherwise>remarkable</xsl:otherwise>
    </xsl:choose>
  </xsl:variable>
  <xsl:variable name="kind" select="/page/@kind"/>
  <xsl:variable name="loggedIn" select="boolean(/page/user)"/>

  <!-- ========== Root ========== -->
  <xsl:template match="/page">
    <html lang="en">
      <head>
        <meta charset="utf-8"/>
        <meta name="viewport" content="width=device-width, initial-scale=1"/>
        <title>
          <xsl:choose>
            <xsl:when test="@title != ''"><xsl:value-of select="@title"/></xsl:when>
            <xsl:otherwise>rmfakecloud</xsl:otherwise>
          </xsl:choose>
        </title>
        <link rel="stylesheet" href="/assets/app.css"/>
        <xsl:if test="theme-css">
          <style id="rm-shell-theme">
            <xsl:value-of select="theme-css" disable-output-escaping="yes"/>
          </style>
        </xsl:if>
      </head>
      <body>
        <xsl:attribute name="class">
          <xsl:text>chrome-</xsl:text>
          <xsl:value-of select="$chrome"/>
          <xsl:text> page-</xsl:text>
          <xsl:value-of select="$kind"/>
        </xsl:attribute>
        <xsl:if test="@path != ''">
          <xsl:attribute name="data-path"><xsl:value-of select="@path"/></xsl:attribute>
        </xsl:if>
        <xsl:if test="$isAdmin">
          <xsl:attribute name="data-admin">true</xsl:attribute>
        </xsl:if>
        <xsl:if test="/page/user/@email != ''">
          <xsl:attribute name="data-email"><xsl:value-of select="/page/user/@email"/></xsl:attribute>
        </xsl:if>
        <xsl:if test="/page/user/@id != ''">
          <xsl:attribute name="data-user"><xsl:value-of select="/page/user/@id"/></xsl:attribute>
        </xsl:if>

        <a class="skip-link" href="#main">Skip to content</a>

        <xsl:if test="nav and $kind != 'login' and $kind != 'error'">
          <xsl:apply-templates select="nav"/>
        </xsl:if>

        <div id="rm-shell-root">
          <xsl:apply-templates select="flash"/>

          <main id="main" class="page-main">
            <xsl:apply-templates select="body"/>
          </main>
        </div>

        <xsl:if test="nav and $kind != 'login' and $kind != 'error'">
          <script type="application/json" id="rm-nav">
            <xsl:text>[</xsl:text>
            <xsl:for-each select="nav/item">
              <xsl:if test="position() &gt; 1"><xsl:text>,</xsl:text></xsl:if>
              <xsl:text>{"id":"</xsl:text><xsl:value-of select="@id"/><xsl:text>",</xsl:text>
              <xsl:text>"href":"</xsl:text><xsl:value-of select="@href"/><xsl:text>",</xsl:text>
              <xsl:text>"label":"</xsl:text><xsl:value-of select="@label"/><xsl:text>",</xsl:text>
              <xsl:text>"icon":"</xsl:text><xsl:value-of select="@icon"/><xsl:text>",</xsl:text>
              <xsl:text>"adminOnly":</xsl:text>
              <xsl:choose>
                <xsl:when test="@admin-only = 'true'"><xsl:text>true</xsl:text></xsl:when>
                <xsl:otherwise><xsl:text>false</xsl:text></xsl:otherwise>
              </xsl:choose>
              <xsl:text>}</xsl:text>
            </xsl:for-each>
            <xsl:text>]</xsl:text>
          </script>
        </xsl:if>

        <xsl:call-template name="page-scripts"/>
      </body>
    </html>
  </xsl:template>

  <!-- ========== Flash ========== -->
  <xsl:template match="flash">
    <div role="status">
      <xsl:attribute name="class">
        <xsl:text>flash flash-</xsl:text>
        <xsl:choose>
          <xsl:when test="@type = 'error' or @type = 'info' or @type = 'success'">
            <xsl:value-of select="@type"/>
          </xsl:when>
          <xsl:otherwise>info</xsl:otherwise>
        </xsl:choose>
      </xsl:attribute>
      <xsl:value-of select="."/>
    </div>
  </xsl:template>

  <!-- ========== Navigation ========== -->
  <xsl:template match="nav">
    <header class="site-header">
      <nav class="site-nav" aria-label="Main">
        <div class="nav-bar">
          <xsl:call-template name="nav-brand"/>
          <ul class="nav-links">
            <xsl:apply-templates select="item"/>
          </ul>
          <xsl:if test="$loggedIn">
            <div class="nav-user-slot">
              <xsl:call-template name="user-menu"/>
            </div>
          </xsl:if>
        </div>
      </nav>
    </header>
  </xsl:template>

  <xsl:template name="nav-brand">
    <a class="nav-brand" href="/">
      <span class="icon icon-cloud" aria-hidden="true"/>
      <span class="brand-text">rmfakecloud</span>
    </a>
  </xsl:template>

  <xsl:template match="nav/item">
    <xsl:if test="not(@admin-only = 'true') or $isAdmin">
      <li>
        <xsl:attribute name="class">
          <xsl:text>nav-item</xsl:text>
          <xsl:if test="@active = 'true'"> nav-item-active</xsl:if>
        </xsl:attribute>
        <a>
          <xsl:attribute name="href"><xsl:value-of select="@href"/></xsl:attribute>
          <xsl:attribute name="class">
            <xsl:text>nav-link</xsl:text>
            <xsl:if test="@active = 'true'"> is-active</xsl:if>
          </xsl:attribute>
          <xsl:if test="@active = 'true'">
            <xsl:attribute name="aria-current">page</xsl:attribute>
          </xsl:if>
          <xsl:if test="@icon != ''">
            <span aria-hidden="true">
              <xsl:attribute name="class">
                <xsl:text>icon icon-</xsl:text>
                <xsl:value-of select="@icon"/>
              </xsl:attribute>
            </span>
          </xsl:if>
          <xsl:value-of select="@label"/>
        </a>
      </li>
    </xsl:if>
  </xsl:template>

  <xsl:template name="user-menu">
    <details class="user-menu">
      <summary>
        <span class="icon icon-person" aria-hidden="true"/>
        <xsl:choose>
          <xsl:when test="/page/user/@email != ''">
            <xsl:value-of select="/page/user/@email"/>
          </xsl:when>
          <xsl:otherwise>
            <xsl:value-of select="/page/user/@id"/>
          </xsl:otherwise>
        </xsl:choose>
      </summary>
      <ul class="user-menu-list">
        <li><a href="/profile">Profile</a></li>
        <li><a href="/help">Help</a></li>
        <xsl:if test="$isAdmin">
          <li><a href="/admin/themes">Themes</a></li>
        </xsl:if>
        <li>
          <form class="inline-form" method="post" action="/logout">
            <button type="submit" class="btn btn-link">Log out</button>
          </form>
        </li>
      </ul>
    </details>
  </xsl:template>

  <!-- ========== Body dispatch ========== -->
  <xsl:template match="body">
    <xsl:apply-templates select="*"/>
  </xsl:template>

  <!-- home -->
  <xsl:template match="home">
    <article class="panel home-panel">
      <h1>rmfakecloud</h1>
      <xsl:apply-templates select="p"/>
    </article>
  </xsl:template>

  <xsl:template match="home/p">
    <p><xsl:value-of select="."/></p>
  </xsl:template>

  <!-- help -->
  <xsl:template match="help">
    <article class="panel help-panel">
      <h1>Help</h1>
      <xsl:apply-templates select="section"/>
    </article>
  </xsl:template>

  <xsl:template match="help/section">
    <section class="help-section">
      <xsl:attribute name="id"><xsl:value-of select="@id"/></xsl:attribute>
      <h2><xsl:value-of select="@title"/></h2>
      <xsl:if test="intro">
        <p class="help-intro"><xsl:value-of select="intro"/></p>
      </xsl:if>
      <ul class="help-links">
        <xsl:apply-templates select="link"/>
      </ul>
    </section>
  </xsl:template>

  <xsl:template match="help/section/link">
    <li>
      <a>
        <xsl:attribute name="href"><xsl:value-of select="@href"/></xsl:attribute>
        <xsl:if test="@internal != 'true'">
          <xsl:attribute name="rel">noopener noreferrer</xsl:attribute>
          <xsl:attribute name="target">_blank</xsl:attribute>
        </xsl:if>
        <xsl:value-of select="."/>
      </a>
      <xsl:if test="@note != ''">
        <span class="help-note"> — <xsl:value-of select="@note"/></span>
      </xsl:if>
    </li>
  </xsl:template>

  <!-- login -->
  <xsl:template match="login">
    <article class="panel login-panel">
      <xsl:if test="@show-brand = 'true'">
        <div class="login-brand">
          <span class="icon icon-cloud icon-lg" aria-hidden="true"/>
          <h1>rmfakecloud</h1>
        </div>
      </xsl:if>
      <xsl:if test="@error != ''">
        <div class="flash flash-error" role="alert"><xsl:value-of select="@error"/></div>
      </xsl:if>
      <form class="login-form" method="post" action="/login" autocomplete="on">
        <div class="field">
          <label for="login-email">Email or user ID</label>
          <input id="login-email" type="text" name="email" required="required" autocomplete="username" autocapitalize="none" spellcheck="false">
            <xsl:if test="@email != ''">
              <xsl:attribute name="value"><xsl:value-of select="@email"/></xsl:attribute>
            </xsl:if>
          </input>
        </div>
        <div class="field">
          <label for="login-password">Password</label>
          <input id="login-password" type="password" name="password" required="required" autocomplete="current-password"/>
        </div>
        <div class="form-actions">
          <xsl:if test="@passkey = 'true'">
            <p id="passkey-error" class="flash flash-error" hidden="hidden"></p>
            <button type="button" class="btn btn-secondary" id="passkey-login" data-action="passkey-login">
              Sign in with passkey
            </button>
          </xsl:if>
          <button type="submit" class="btn btn-primary">Login</button>
        </div>
      </form>
      <xsl:if test="@registration = 'true'">
        <p class="login-register">
          <a href="/register">Create an account</a>
        </p>
      </xsl:if>
    </article>
  </xsl:template>

  <!-- connect -->
  <xsl:template match="connect">
    <article class="panel connect-panel">
      <h1 class="visually-hidden">Connect</h1>
      <div class="connect-stack">
        <button type="button" class="btn btn-secondary" id="connect-refresh" aria-label="Generate new code">
          Refresh code
        </button>

        <xsl:variable name="promptLoc">
          <xsl:choose>
            <xsl:when test="@prompt-location != ''"><xsl:value-of select="@prompt-location"/></xsl:when>
            <xsl:otherwise>above</xsl:otherwise>
          </xsl:choose>
        </xsl:variable>
        <xsl:variable name="promptStyle">
          <xsl:choose>
            <xsl:when test="@prompt-style != ''"><xsl:value-of select="@prompt-style"/></xsl:when>
            <xsl:otherwise>plain</xsl:otherwise>
          </xsl:choose>
        </xsl:variable>
        <xsl:variable name="promptText">
          <xsl:choose>
            <xsl:when test="@prompt != ''"><xsl:value-of select="@prompt"/></xsl:when>
            <xsl:otherwise>Enter this code on your tablet</xsl:otherwise>
          </xsl:choose>
        </xsl:variable>

        <div class="connect-main">
          <xsl:attribute name="data-prompt-location"><xsl:value-of select="$promptLoc"/></xsl:attribute>

          <xsl:if test="$promptLoc = 'above'">
            <p>
              <xsl:attribute name="class">
                <xsl:text>connect-prompt prompt-</xsl:text>
                <xsl:value-of select="$promptStyle"/>
              </xsl:attribute>
              <xsl:value-of select="$promptText"/>
            </p>
          </xsl:if>

          <div class="connect-code" id="connect-code" data-code="{@code}">
            <xsl:choose>
              <xsl:when test="@code != ''"><xsl:value-of select="@code"/></xsl:when>
              <xsl:otherwise>· · · · · · · ·</xsl:otherwise>
            </xsl:choose>
          </div>
          <p id="connect-status" class="connect-status" role="status" aria-live="polite" hidden="hidden"/>


          <xsl:if test="$promptLoc = 'below'">
            <p>
              <xsl:attribute name="class">
                <xsl:text>connect-prompt prompt-</xsl:text>
                <xsl:value-of select="$promptStyle"/>
              </xsl:attribute>
              <xsl:value-of select="$promptText"/>
            </p>
          </xsl:if>
        </div>

        <xsl:variable name="gaugeType">
          <xsl:choose>
            <xsl:when test="@gauge-type != ''"><xsl:value-of select="@gauge-type"/></xsl:when>
            <xsl:otherwise>circle</xsl:otherwise>
          </xsl:choose>
        </xsl:variable>

        <div class="connect-gauge" id="connect-gauge" role="progressbar" aria-valuemin="0" aria-valuemax="100" aria-valuenow="100" data-ttl="300">
          <xsl:attribute name="data-gauge-type"><xsl:value-of select="$gaugeType"/></xsl:attribute>
          <!-- Gauge markup built by connect.js so animation owns a single code path -->
        </div>
      </div>
    </article>
  </xsl:template>

  <!-- profile -->
  <xsl:template match="profile">
    <article class="panel profile-panel">
      <h1>Profile</h1>

      <xsl:if test="themes">
        <section class="profile-themes" aria-labelledby="theme-deploy-heading">
          <h2 id="theme-deploy-heading">Deploy theme</h2>
          <form method="post" action="/profile/theme" class="theme-deploy-form">
            <div class="field">
              <label for="profile-theme-id">Published theme</label>
              <select id="profile-theme-id" name="themeId" required="required">
                <xsl:for-each select="themes/theme">
                  <option value="{@id}">
                    <xsl:if test="@selected = 'true'">
                      <xsl:attribute name="selected">selected</xsl:attribute>
                    </xsl:if>
                    <xsl:value-of select="@name"/>
                  </option>
                </xsl:for-each>
              </select>
            </div>
            <xsl:if test="overrides/color">
              <fieldset class="color-overrides">
                <legend>Personal color overrides</legend>
                <xsl:for-each select="overrides/color">
                  <div class="field field-inline">
                    <label>
                      <xsl:attribute name="for">
                        <xsl:text>override-</xsl:text>
                        <xsl:value-of select="@key"/>
                      </xsl:attribute>
                      <xsl:value-of select="@key"/>
                    </label>
                    <input type="color" name="override-{@key}" value="{@value}">
                      <xsl:attribute name="id">
                        <xsl:text>override-</xsl:text>
                        <xsl:value-of select="@key"/>
                      </xsl:attribute>
                    </input>
                  </div>
                </xsl:for-each>
              </fieldset>
            </xsl:if>
            <div class="form-actions">
              <button type="submit" class="btn btn-primary">Deploy</button>
            </div>
          </form>
        </section>
      </xsl:if>

      <xsl:if test="passkeys">
        <section class="profile-passkeys" aria-labelledby="passkeys-heading">
          <h2 id="passkeys-heading">Passkeys</h2>
          <xsl:choose>
            <xsl:when test="passkeys/@enabled = 'true'">
              <table class="data-table">
                <thead>
                  <tr>
                    <th scope="col">Name</th>
                    <th scope="col">Created</th>
                    <th scope="col"><span class="visually-hidden">Actions</span></th>
                  </tr>
                </thead>
                <tbody>
                  <xsl:for-each select="passkeys/cred">
                    <tr>
                      <td><xsl:value-of select="@name"/></td>
                      <td><xsl:value-of select="@created"/></td>
                      <td>
                        <form method="post" action="/profile/passkeys/{@id}/delete" class="inline-form">
                          <button type="submit" class="btn btn-danger btn-sm" data-cred-delete="{@id}">Delete</button>
                        </form>
                      </td>
                    </tr>
                  </xsl:for-each>
                  <xsl:if test="not(passkeys/cred)">
                    <tr><td colspan="3">No passkeys registered.</td></tr>
                  </xsl:if>
                </tbody>
              </table>
              <p id="passkey-error" class="flash flash-error" hidden="hidden"></p>
              <button type="button" class="btn btn-secondary" id="passkey-register" data-action="passkey-register">
                Register passkey
              </button>
            </xsl:when>
            <xsl:otherwise>
              <p class="muted">Passkeys are not enabled on this server.</p>
            </xsl:otherwise>
          </xsl:choose>
        </section>
      </xsl:if>

      <section class="profile-password" aria-labelledby="password-heading">
        <h2 id="password-heading">Change password</h2>
        <form method="post" action="/profile/password" class="password-form" autocomplete="off">
          <div class="field">
            <label for="current-password">Current password</label>
            <input id="current-password" type="password" name="currentPassword" required="required" autocomplete="current-password"/>
          </div>
          <div class="field">
            <label for="new-password-profile">New password</label>
            <input id="new-password-profile" type="password" name="newPassword" required="required" autocomplete="new-password"/>
          </div>
          <div class="form-actions">
            <button type="submit" class="btn btn-primary">Update password</button>
          </div>
        </form>
      </section>
    </article>
  </xsl:template>

  <!-- documents -->
  <xsl:template match="documents">
    <article class="panel documents-panel rm-files">
      <header class="rm-files-header">
        <div class="rm-files-title-row">
          <xsl:if test="@folder-id != ''">
            <p class="rm-files-crumb">
              <a href="/documents">My Files</a>
              <xsl:if test="@parent-id != ''">
                <xsl:text> / </xsl:text>
                <a href="/documents?folder={@parent-id}">Parent</a>
              </xsl:if>
            </p>
          </xsl:if>
          <h1>
            <xsl:choose>
              <xsl:when test="@folder-name != ''"><xsl:value-of select="@folder-name"/></xsl:when>
              <xsl:otherwise>My Files</xsl:otherwise>
            </xsl:choose>
          </h1>
        </div>
        <div class="rm-files-sort">
          <label for="rm-sort">Sort</label>
          <select id="rm-sort">
            <option value="modified" selected="selected">Last modified</option>
            <option value="name">Name</option>
          </select>
        </div>
      </header>

      <div id="rm-search-bar" class="rm-search-bar" hidden="hidden">
        <label for="rm-search">Search files</label>
        <input id="rm-search" type="search" name="q" autocomplete="off"/>
      </div>

      <section class="rm-folder-cluster" aria-label="Folders">
        <ul class="rm-folder-grid">
          <xsl:for-each select="folders/folder">
            <li class="rm-folder-item" data-name="{@name}" data-modified="{@modified}">
              <a href="/documents?folder={@id}">
                <span class="rm-folder-icon" aria-hidden="true"></span>
                <span class="rm-folder-name"><xsl:value-of select="@name"/></span>
                <xsl:if test="@pinned = 'true'">
                  <span class="rm-star" aria-label="Favorite">★</span>
                </xsl:if>
              </a>
            </li>
          </xsl:for-each>
        </ul>
      </section>

      <section class="rm-file-cluster" aria-label="Notebooks and documents">
        <ul class="rm-file-grid">
          <xsl:for-each select="files/doc">
            <li class="rm-file-item" data-name="{@name}" data-modified="{@modified}" data-type="{@type}">
              <a href="/documents/{@id}" class="rm-file-link">
                <span class="rm-page-frame is-{@type}">
                  <xsl:choose>
                    <xsl:when test="@type = 'pdf'">
                      <canvas class="rm-thumb-canvas" width="180" height="240" data-pdf-url="/ui/api/documents/{@id}?type=pdf" data-pdf-page="{@thumb-page}" aria-hidden="true"></canvas>
                    </xsl:when>
                    <xsl:when test="@type = 'notebook'">
                      <img class="rm-thumb-img" src="/ui/api/documents/{@id}/page/{@thumb-page}" alt="" loading="lazy" decoding="async" width="180" height="240"/>
                    </xsl:when>
                    <xsl:when test="@type = 'epub'">
                      <img class="rm-thumb-img" src="/ui/api/documents/{@id}/epub/thumb" alt="" loading="lazy" decoding="async" width="180" height="240"/>
                    </xsl:when>
                    <xsl:otherwise>
                      <span class="rm-page-placeholder" aria-hidden="true"></span>
                    </xsl:otherwise>
                  </xsl:choose>
                  <xsl:if test="@type != 'epub'">
                    <span class="rm-page-ear" aria-hidden="true"></span>
                  </xsl:if>
                </span>
                <span class="rm-file-meta">
                  <span class="rm-file-name">
                    <xsl:value-of select="@name"/>
                    <xsl:if test="@pinned = 'true'">
                      <span class="rm-star" aria-label="Favorite">★</span>
                    </xsl:if>
                  </span>
                  <span class="rm-file-sub">
                    <xsl:choose>
                      <xsl:when test="@pages != '' and number(@pages) &gt; 0">
                        <xsl:text>Page </xsl:text>
                        <xsl:value-of select="number(@page) + 1"/>
                        <xsl:text> of </xsl:text>
                        <xsl:value-of select="@pages"/>
                      </xsl:when>
                      <xsl:when test="@type = 'pdf'">PDF</xsl:when>
                      <xsl:when test="@type = 'epub'">EPUB</xsl:when>
                      <xsl:otherwise>Notebook</xsl:otherwise>
                    </xsl:choose>
                  </span>
                </span>
              </a>
            </li>
          </xsl:for-each>
        </ul>
        <xsl:if test="not(files/doc) and not(folders/folder)">
          <p class="rm-files-empty">This folder is empty.</p>
        </xsl:if>
      </section>

      <div class="rm-dock" role="toolbar" aria-label="File actions">
        <button type="button" id="rm-search-toggle" aria-label="Search" aria-expanded="false" aria-controls="rm-search-bar">
          <span class="rm-dock-icon rm-dock-search" aria-hidden="true"></span>
        </button>
        <button type="button" id="rm-upload-toggle" aria-label="Upload document">
          <span class="rm-dock-icon rm-dock-plus" aria-hidden="true"></span>
        </button>
        <button type="button" id="rm-folder-toggle" aria-label="New folder" aria-haspopup="dialog" aria-controls="rm-folder-dialog">
          <span class="rm-dock-icon rm-dock-folder" aria-hidden="true"></span>
        </button>
      </div>

      <form id="rm-upload-form" method="post" action="/documents/upload" enctype="multipart/form-data" class="visually-hidden">
        <xsl:if test="@folder-id != ''">
          <input type="hidden" name="parent" value="{@folder-id}"/>
        </xsl:if>
        <label for="doc-upload">Upload file</label>
        <input id="doc-upload" type="file" name="file"/>
      </form>

      <dialog id="rm-folder-dialog" class="rm-folder-dialog">
        <form method="post" action="/documents/folder">
          <xsl:if test="@folder-id != ''">
            <input type="hidden" name="parent" value="{@folder-id}"/>
          </xsl:if>
          <h2>New folder</h2>
          <div class="field">
            <label for="folder-name">Folder name</label>
            <input id="folder-name" type="text" name="name" required="required" maxlength="200"/>
          </div>
          <div class="form-actions">
            <button type="button" class="btn btn-secondary" id="rm-folder-cancel">Cancel</button>
            <button type="submit" class="btn btn-primary">Create</button>
          </div>
        </form>
      </dialog>
    </article>
  </xsl:template>

  <xsl:template match="tree/folder|folder/folder">
    <li class="tree-folder">
      <a href="/documents?folder={@id}">
        <span class="icon icon-folder" aria-hidden="true"/>
        <xsl:value-of select="@name"/>
      </a>
      <xsl:if test="folder|doc">
        <ul>
          <xsl:apply-templates select="folder|doc"/>
        </ul>
      </xsl:if>
    </li>
  </xsl:template>

  <xsl:template match="tree/doc|folder/doc">
    <li class="tree-doc">
      <a href="/documents/{@id}">
        <span class="icon icon-file" aria-hidden="true"/>
        <xsl:value-of select="@name"/>
      </a>
    </li>
  </xsl:template>

  <xsl:template match="doc" mode="file-row">
    <tr>
      <td>
        <xsl:choose>
          <xsl:when test="@type = 'folder'">
            <a href="/documents?folder={@id}"><xsl:value-of select="@name"/></a>
          </xsl:when>
          <xsl:otherwise>
            <a href="/documents/{@id}"><xsl:value-of select="@name"/></a>
          </xsl:otherwise>
        </xsl:choose>
      </td>
      <td><xsl:value-of select="@type"/></td>
      <td><xsl:value-of select="@size"/></td>
      <td><xsl:value-of select="@modified"/></td>
      <td class="actions">
        <xsl:if test="@type != 'folder'">
          <a class="btn btn-secondary btn-sm" href="/ui/api/documents/{@id}?type=pdf">PDF</a>
        </xsl:if>
        <form method="post" action="/documents/{@id}/delete" class="inline-form">
          <button type="submit" class="btn btn-danger btn-sm">Delete</button>
        </form>
      </td>
    </tr>
  </xsl:template>

  <!-- integrations -->
  <xsl:template match="integrations">
    <article class="panel integrations-panel">
      <h1>Integrations</h1>
      <table class="data-table">
        <thead>
          <tr>
            <th scope="col">Name</th>
            <th scope="col">Provider</th>
            <th scope="col"><span class="visually-hidden">Actions</span></th>
          </tr>
        </thead>
        <tbody>
          <xsl:for-each select="integration">
            <tr>
              <td><xsl:value-of select="@name"/></td>
              <td><xsl:value-of select="@provider"/></td>
              <td>
                <form method="post" action="/integrations/{@id}/delete" class="inline-form">
                  <button type="submit" class="btn btn-danger btn-sm">Delete</button>
                </form>
              </td>
            </tr>
          </xsl:for-each>
          <xsl:if test="not(integration)">
            <tr><td colspan="3">No integrations yet.</td></tr>
          </xsl:if>
        </tbody>
      </table>

      <section class="create-integration" aria-labelledby="new-integration-heading">
        <h2 id="new-integration-heading">Add integration</h2>
        <form method="post" action="/integrations" class="integration-form" autocomplete="off">
          <div class="field">
            <label for="integration-name">Name</label>
            <input id="integration-name" type="text" name="name" required="required"/>
          </div>
          <div class="field">
            <label for="integration-provider">Provider</label>
            <select id="integration-provider" name="provider" required="required">
              <option value="localfs">Local filesystem</option>
              <option value="webdav">WebDAV</option>
              <option value="ftp">FTP</option>
              <option value="dropbox">Dropbox</option>
              <option value="google">Google Drive</option>
              <option value="webhook">Webhook</option>
            </select>
          </div>
          <div class="form-actions">
            <button type="submit" class="btn btn-primary">Create</button>
          </div>
        </form>
      </section>
    </article>
  </xsl:template>

  <!-- admin -->
  <xsl:template match="admin">
    <article class="panel admin-panel">
      <h1>Admin</h1>
      <p><a href="/admin/themes">Open Theme studio</a></p>
      <table class="data-table">
        <thead>
          <tr>
            <th scope="col">User ID</th>
            <th scope="col">Email</th>
            <th scope="col">Name</th>
            <th scope="col">Admin</th>
            <th scope="col"><span class="visually-hidden">Actions</span></th>
          </tr>
        </thead>
        <tbody>
          <xsl:for-each select="user">
            <tr>
              <td><xsl:value-of select="@id"/></td>
              <td><xsl:value-of select="@email"/></td>
              <td><xsl:value-of select="@name"/></td>
              <td>
                <xsl:choose>
                  <xsl:when test="@admin = 'true'">yes</xsl:when>
                  <xsl:otherwise>no</xsl:otherwise>
                </xsl:choose>
              </td>
              <td>
                <form method="post" action="/admin/users/{@id}/delete" class="inline-form">
                  <button type="submit" class="btn btn-danger btn-sm">Delete</button>
                </form>
              </td>
            </tr>
          </xsl:for-each>
          <xsl:if test="not(user)">
            <tr><td colspan="5">No users.</td></tr>
          </xsl:if>
        </tbody>
      </table>

      <section class="create-user" aria-labelledby="new-user-heading">
        <h2 id="new-user-heading">Create user</h2>
        <form method="post" action="/admin/users" class="user-form" autocomplete="off">
          <div class="field">
            <label for="new-userid">User ID</label>
            <input id="new-userid" type="text" name="userid" required="required"/>
          </div>
          <div class="field">
            <label for="new-email">Email</label>
            <input id="new-email" type="email" name="email" required="required"/>
          </div>
          <div class="field">
            <label for="new-password">Password</label>
            <input id="new-password" type="password" name="newPassword" required="required" autocomplete="new-password"/>
          </div>
          <div class="form-actions">
            <button type="submit" class="btn btn-primary">Create user</button>
          </div>
        </form>
      </section>
    </article>
  </xsl:template>

  <!-- themes studio -->
  <xsl:template match="themes-studio">
    <article class="panel themes-panel">
      <h1>Theme studio</h1>
      <p><a href="/admin">Back to admin</a></p>

      <section aria-labelledby="themes-list-heading">
        <h2 id="themes-list-heading">Themes</h2>
        <table class="data-table">
          <thead>
            <tr>
              <th scope="col">ID</th>
              <th scope="col">Name</th>
              <th scope="col">Published</th>
              <th scope="col">Built-in</th>
              <th scope="col"><span class="visually-hidden">Actions</span></th>
            </tr>
          </thead>
          <tbody>
            <xsl:for-each select="theme">
              <tr>
                <td><xsl:value-of select="@id"/></td>
                <td><xsl:value-of select="@name"/></td>
                <td>
                  <xsl:choose>
                    <xsl:when test="@published = 'true'">yes</xsl:when>
                    <xsl:otherwise>no</xsl:otherwise>
                  </xsl:choose>
                </td>
                <td>
                  <xsl:choose>
                    <xsl:when test="@builtin = 'true'">yes</xsl:when>
                    <xsl:otherwise>no</xsl:otherwise>
                  </xsl:choose>
                </td>
                <td class="actions">
                  <a class="btn btn-secondary btn-sm" href="/admin/themes?id={@id}">Edit</a>
                  <xsl:if test="@builtin != 'true'">
                    <form method="post" action="/admin/themes/{@id}/publish" class="inline-form">
                      <xsl:choose>
                        <xsl:when test="@published = 'true'">
                          <input type="hidden" name="published" value="false"/>
                          <button type="submit" class="btn btn-secondary btn-sm">Unpublish</button>
                        </xsl:when>
                        <xsl:otherwise>
                          <input type="hidden" name="published" value="true"/>
                          <button type="submit" class="btn btn-secondary btn-sm">Publish</button>
                        </xsl:otherwise>
                      </xsl:choose>
                    </form>
                    <form method="post" action="/admin/themes/{@id}/delete" class="inline-form">
                      <button type="submit" class="btn btn-danger btn-sm">Delete</button>
                    </form>
                  </xsl:if>
                </td>
              </tr>
            </xsl:for-each>
          </tbody>
        </table>
      </section>

      <xsl:if test="editor">
        <section class="theme-editor" aria-labelledby="theme-editor-heading">
          <h2 id="theme-editor-heading">
            <xsl:choose>
              <xsl:when test="editor/@id != ''">Edit theme</xsl:when>
              <xsl:otherwise>New theme</xsl:otherwise>
            </xsl:choose>
          </h2>
          <form method="post" action="/admin/themes/save" class="theme-save-form" id="theme-studio-form">
            <div class="field">
              <label for="theme-id">ID</label>
              <input id="theme-id" type="text" name="id" required="required" value="{editor/@id}"/>
            </div>
            <div class="field">
              <label for="theme-name">Name</label>
              <input id="theme-name" type="text" name="name" required="required" value="{editor/@name}"/>
            </div>
            <div class="field field-checkbox">
              <input id="theme-published" type="checkbox" name="published" value="true">
                <xsl:if test="editor/@published = 'true'">
                  <xsl:attribute name="checked">checked</xsl:attribute>
                </xsl:if>
              </input>
              <label for="theme-published">Published</label>
            </div>
            <div class="field">
              <label for="theme-xml">Theme XML</label>
              <p class="muted" id="chrome-style-help">
                Chrome layout styles for <code>layout/chrome/@style</code> and
                <code>layout/mobile/chrome/@style</code>:
                <code>remarkable</code>, <code>googledocs</code>, <code>icloud</code>, or <code>os</code>.
                Unknown values fall back to <code>remarkable</code>.
              </p>
              <textarea id="theme-xml" name="xml" rows="24" cols="80" spellcheck="false" aria-describedby="chrome-style-help">
                <xsl:value-of select="editor" disable-output-escaping="yes"/>
              </textarea>
            </div>
            <div class="form-actions">
              <button type="submit" class="btn btn-primary">Save</button>
            </div>
          </form>
        </section>
      </xsl:if>
    </article>
  </xsl:template>

  <!-- screenshare -->
  <xsl:template match="screenshare">
    <article class="panel screenshare-panel">
      <h1>Screen share</h1>
      <p id="ss-status" class="muted" role="status">Waiting for reMarkable to start screen sharing…</p>
      <div id="ss-stage" class="screenshare-layout">
        <canvas id="ss-canvas" width="1404" height="1872" aria-label="Device screen"></canvas>
        <div id="pen-cursor" hidden="hidden"></div>
      </div>
      <div id="ss-controls" class="screenshare-controls">
        <button type="button" class="btn btn-secondary btn-sm" id="ss-rotate-l" title="Rotate left">Rotate L</button>
        <button type="button" class="btn btn-secondary btn-sm" id="ss-rotate-r" title="Rotate right">Rotate R</button>
        <button type="button" class="btn btn-secondary btn-sm" id="ss-fullscreen">Fullscreen</button>
        <button type="button" class="btn btn-secondary btn-sm" id="ss-exit-fs" hidden="hidden">Exit fullscreen</button>
        <button type="button" class="btn btn-secondary btn-sm" id="ss-options">Options</button>
        <button type="button" class="btn btn-danger btn-sm" id="ss-disconnect">Disconnect</button>
        <button type="button" class="btn btn-primary btn-sm" id="ss-reconnect" hidden="hidden">Reconnect</button>
      </div>
      <div id="ss-options-panel" class="screenshare-options" hidden="hidden">
        <p>Backdrop</p>
        <button type="button" data-backdrop="white" title="white">White</button>
        <button type="button" data-backdrop="off-white" title="off-white">Off-white</button>
        <button type="button" data-backdrop="gray" title="gray">Gray</button>
        <button type="button" data-backdrop="black" title="black">Black</button>
        <label for="ss-custom-color">Custom
          <input id="ss-custom-color" type="color" value="#808080"/>
        </label>
      </div>
    </article>
  </xsl:template>

  <!-- pdf -->
  <xsl:template match="pdf">
    <article class="panel pdf-panel">
      <header class="pdf-header">
        <h1>
          <xsl:choose>
            <xsl:when test="@name != ''"><xsl:value-of select="@name"/></xsl:when>
            <xsl:otherwise>PDF</xsl:otherwise>
          </xsl:choose>
        </h1>
        <p>
          <a href="/documents">Back to documents</a>
          <xsl:text> · </xsl:text>
          <a>
            <xsl:attribute name="href"><xsl:value-of select="@url"/></xsl:attribute>
            Download PDF
          </a>
        </p>
      </header>
      <div
        id="pdf-viewer"
        class="pdf-viewer"
        data-doc-id="{@doc-id}"
        data-doc-url="{@url}"
      >
        <div id="pdf-canvas-container"></div>
      </div>
    </article>
  </xsl:template>

  <!-- epub as a website -->
  <xsl:template match="epub">
    <article class="panel epub-panel">
      <header class="epub-header">
        <h1>
          <xsl:choose>
            <xsl:when test="@name != ''"><xsl:value-of select="@name"/></xsl:when>
            <xsl:otherwise>EPUB</xsl:otherwise>
          </xsl:choose>
        </h1>
        <p>
          <a href="/documents">Back to documents</a>
          <xsl:if test="@download-href != ''">
            <xsl:text> · </xsl:text>
            <a>
              <xsl:attribute name="href"><xsl:value-of select="@download-href"/></xsl:attribute>
              <xsl:attribute name="download"/>
              Download EPUB
            </a>
          </xsl:if>
        </p>
      </header>
      <div class="epub-layout">
        <nav class="epub-toc" aria-label="Contents">
          <h2>Contents</h2>
          <ol>
            <xsl:for-each select="spine/item">
              <li>
                <a target="epub-frame">
                  <xsl:attribute name="href"><xsl:value-of select="@href"/></xsl:attribute>
                  <xsl:attribute name="title"><xsl:value-of select="@path"/></xsl:attribute>
                  <xsl:if test="@href = /page/body/epub/@start-href">
                    <xsl:attribute name="aria-current">page</xsl:attribute>
                  </xsl:if>
                  <xsl:value-of select="@label"/>
                </a>
              </li>
            </xsl:for-each>
          </ol>
        </nav>
        <div class="epub-stage">
          <xsl:choose>
            <xsl:when test="@start-href != ''">
              <iframe
                id="epub-frame"
                name="epub-frame"
                class="epub-frame"
                referrerpolicy="same-origin"
              >
                <xsl:attribute name="title">
                  <xsl:choose>
                    <xsl:when test="@name != ''"><xsl:value-of select="@name"/></xsl:when>
                    <xsl:otherwise>EPUB</xsl:otherwise>
                  </xsl:choose>
                </xsl:attribute>
                <xsl:attribute name="src"><xsl:value-of select="@start-href"/></xsl:attribute>
                <xsl:comment>epub website</xsl:comment>
              </iframe>
            </xsl:when>
            <xsl:otherwise>
              <p class="muted">This EPUB could not be opened as a website.</p>
            </xsl:otherwise>
          </xsl:choose>
        </div>
      </div>
    </article>
  </xsl:template>

  <!-- error -->
  <xsl:template match="error">
    <article class="panel error-panel">
      <h1>
        <xsl:choose>
          <xsl:when test="@code != ''"><xsl:value-of select="@code"/></xsl:when>
          <xsl:otherwise>Error</xsl:otherwise>
        </xsl:choose>
      </h1>
      <p>
        <xsl:choose>
          <xsl:when test="@message != ''"><xsl:value-of select="@message"/></xsl:when>
          <xsl:otherwise>Something went wrong.</xsl:otherwise>
        </xsl:choose>
      </p>
      <p><a href="/">Home</a></p>
    </article>
  </xsl:template>

  <!-- Fallback for unknown body children -->
  <xsl:template match="body/*" priority="-1">
    <div class="panel">
      <p class="muted">Unsupported page content.</p>
    </div>
  </xsl:template>

  <!-- ========== Scripts ========== -->
  <xsl:template name="page-scripts">
    <!-- Chrome stack (authenticated pages) -->
    <xsl:if test="$loggedIn and $kind != 'login' and $kind != 'error'">
      <script src="/assets/js/rm.js"></script>
      <script src="/assets/js/legacy/constants.js"></script>
      <script src="/assets/js/legacy/api.service.js"></script>
      <script src="/assets/js/legacy/xslt.js"></script>
      <script src="/assets/js/legacy/auth.js"></script>
      <script src="/assets/js/legacy/theme.js"></script>
      <script src="/assets/js/legacy/useFetch.js"></script>
      <script src="/assets/js/legacy/toast.js"></script>
      <script src="/assets/js/legacy/link.js"></script>
      <script src="/assets/js/legacy/pagination.js"></script>
      <script src="/assets/js/legacy/index.js"></script>
      <script src="/assets/js/shells/all.js"></script>
      <script src="/assets/js/chrome.js"></script>
      <script src="/assets/js/documents-refresh.js"></script>
    </xsl:if>

    <xsl:choose>
      <xsl:when test="$kind = 'login'">
        <xsl:if test="body/login/@passkey = 'true'">
          <script src="/assets/js/vendor/simplewebauthn-browser.umd.min.js" defer="defer"></script>
          <script src="/assets/js/webauthn.js" defer="defer"></script>
        </xsl:if>
      </xsl:when>
      <xsl:when test="$kind = 'profile'">
        <xsl:if test="body/profile/passkeys/@enabled = 'true'">
          <script src="/assets/js/vendor/simplewebauthn-browser.umd.min.js" defer="defer"></script>
          <script src="/assets/js/webauthn.js" defer="defer"></script>
        </xsl:if>
      </xsl:when>
      <xsl:when test="$kind = 'connect'">
        <script src="/assets/js/connect.js" defer="defer"></script>
      </xsl:when>
      <xsl:when test="$kind = 'screenshare'">
        <script src="https://cdn.jsdelivr.net/npm/pako@2.1.0/dist/pako.min.js" defer="defer"></script>
        <script src="/assets/js/screenshare.js" defer="defer"></script>
      </xsl:when>
      <xsl:when test="$kind = 'pdf'">
        <script src="/assets/js/pdfview.js" defer="defer"></script>
      </xsl:when>
      <xsl:when test="$kind = 'epub'">
        <script src="/assets/js/epubview.js" defer="defer"></script>
      </xsl:when>
      <xsl:when test="$kind = 'themes'">
        <script src="/assets/js/theme-studio.js" defer="defer"></script>
      </xsl:when>
    </xsl:choose>
  </xsl:template>

</xsl:stylesheet>
